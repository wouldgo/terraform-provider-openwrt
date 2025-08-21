package provider_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/provider"
	"github.com/foxboron/terraform-provider-openwrt/mocks"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// func testAccPreCheck(t *testing.T) {
// 	// You can add code here to run prior to any test case execution, for example assertions
// 	// about the appropriate environment variables being set are common to see in a pre-check
// 	// function.
// }

func testAccFactories(clientFactory *mocks.MockClientFactory) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"openwrt": providerserver.NewProtocol6WithError(provider.New("test", clientFactory)()),
	}
}

/*

// StateCheck defines an interface for implementing test logic that checks a state file and then returns an error
// if the state file does not match what is expected.
type StateCheck interface {
	// CheckState should perform the state check.
	CheckState(context.Context, CheckStateRequest, *CheckStateResponse)
}
*/

var (
	_ plancheck.PlanCheck   = (*createActionPlanChecker)(nil)
	_ statecheck.StateCheck = (*stateChecker)(nil)
)

type createActionPlanChecker struct {
	addr string
}

func (pC *createActionPlanChecker) CheckPlan(_ context.Context, checkReq plancheck.CheckPlanRequest, checkResponse *plancheck.CheckPlanResponse) {
	raiseErr := true
	for _, aResource := range checkReq.Plan.ResourceChanges {
		if aResource.Address == pC.addr && slices.Contains(aResource.Change.Actions, tfjson.ActionCreate) {
			raiseErr = false
		}
	}
	if raiseErr {
		checkResponse.Error = fmt.Errorf("missing create action")
	}
}

type stateChecker struct {
	addr string
}

func (cS *stateChecker) CheckState(_ context.Context, checkReq statecheck.CheckStateRequest, checkResponse *statecheck.CheckStateResponse) {
	raiseErr := true
	for _, aResource := range checkReq.State.Values.RootModule.Resources {
		if aResource.Address == cS.addr {
			for aKey, aValue := range aResource.AttributeValues {
				typecastedValue, ok := aValue.([]interface{})
				if aKey == "packages" && ok && slices.Contains(typecastedValue, "curl") {
					raiseErr = false
				}
			}
		}
	}

	if raiseErr {
		checkResponse.Error = fmt.Errorf("state is incosistent: %+v", checkReq.State)
	}
}

func mockProviderInit(t *testing.T, client *mocks.MockClient) {
	client.
		EXPECT().
		Auth(gomock.Any(), "root", "test").
		Return(nil).
		Times(6).
		Do(func(_ context.Context, username, password string) {
			t.Logf("Auth method called with: %s, %s", username, password)
		})

	client.
		EXPECT().
		UpdatePackages(gomock.Any()).
		Return(nil).
		Times(6).
		Do(func(_ context.Context) {
			t.Logf("UpdatePackages method called")
		})
}

/*
One dependency missing
Setup: System has everything except one package in the packages list.
Expectation: Plan shows exactly that missing package scheduled for installation, state reflects only that package being changed.

Extra packages present on the system (but not listed in config)
Setup: System has more than packages specifies (e.g. user installed extras manually).
Expectation: Provider should not remove them (unless that’s explicitly part of your semantics). State remains stable.

Empty packages list
Setup: Config has packages = [].
Expectation: Provider does nothing, regardless of what’s installed.

Duplicate entries in packages
Setup: Config accidentally lists the same package twice.
Expectation: Provider de-dupes internally, installs it once.

Invalid package name
Setup: Config has a package name that doesn’t exist in the system’s package manager.
Expectation: Plan should fail gracefully (clear error surfaced, no partial install).

Mixed state (some installed, some missing, some invalid)
Setup: A mix of valid-installed, valid-missing, and invalid package names.
Expectation: Depends on your design choice:

Fail fast with error, or
Install what’s valid, flag invalid ones.
Either way, test that behavior is consistent.

*/

// All dependencies missing
// Setup: System has none of the listed packages.
// Expectation: Plan shows all of them queued for installation, state reflects all being installed.
func TestAccOpkg_AllDepsAreMissing(t *testing.T) {
	os.Setenv("TF_ACC", "1")
	defer os.Unsetenv("TF_ACC")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientFactory := mocks.NewMockClientFactory(ctrl)
	testAccProtoV6ProviderFactories := testAccFactories(clientFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					client := mocks.NewMockClient(ctrl)
					mockProviderInit(t, client)
					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						Return(client, nil).
						AnyTimes()

					gomock.InOrder(
						client.
							EXPECT().
							CheckPackage(gomock.Any(), "curl").
							Return(&api.PackageInfo{
								Version: "",
								Status: api.Status{
									Installed: false,
								},
							}, nil).
							Times(1),
						client.
							EXPECT().
							InstallPackages(gomock.Any(), "curl").
							Return(nil).
							Times(1),

						client.
							EXPECT().
							CheckPackage(gomock.Any(), "curl").
							Return(&api.PackageInfo{
								Version: "test",
								Status: api.Status{
									Installed: true,
								},
							}, nil).
							AnyTimes(),
					)

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						Return(nil).
						MaxTimes(1)
				},
				Config: `
				provider "openwrt" {
					 user = "root"
						password = "test"
						remote = "http://test.lan:8080"
				}
        resource "openwrt_opkg" "test" {
          packages = ["curl"]
        }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_opkg.test", "packages.0", "curl"),
					func(s *terraform.State) error {
						t.Logf("%+v", s)
						return nil
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&createActionPlanChecker{
							addr: "openwrt_opkg.test",
						},
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					&stateChecker{
						addr: "openwrt_opkg.test",
					},
				},
			},
		},
	})
}

// No dependency missing
// Setup: System already has all packages.
// Expectation: Plan shows no changes, state is a no-op.
func TestAccOpkg_NoDepsAreMissing(t *testing.T) {
	os.Setenv("TF_ACC", "1")
	defer os.Unsetenv("TF_ACC")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientFactory := mocks.NewMockClientFactory(ctrl)
	testAccProtoV6ProviderFactories := testAccFactories(clientFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					client := mocks.NewMockClient(ctrl)
					mockProviderInit(t, client)

					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						Return(client, nil).
						AnyTimes()

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(&api.PackageInfo{
							Version: "test",
							Status: api.Status{
								Installed: true,
							},
						}, nil).
						AnyTimes()

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						Return(nil).
						MaxTimes(1)
				},
				Config: `
				provider "openwrt" {
					 user = "root"
						password = "test"
						remote = "http://test.lan:8080"
				}
        resource "openwrt_opkg" "test" {
          packages = ["curl"]
        }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_opkg.test", "packages.0", "curl"),
					func(s *terraform.State) error {
						t.Logf("%+v", s)
						return nil
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						// &createActionPlanChecker{
						// 	addr: "openwrt_opkg.test",
						// },
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					&stateChecker{},
				},
			},
		},
	})
}
