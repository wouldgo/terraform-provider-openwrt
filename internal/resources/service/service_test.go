// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/testutil"

	"github.com/foxboron/terraform-provider-openwrt/mocks"
	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccService_CheckServiceEnabledIfOmitted(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientFactory := mocks.NewMockClientFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(clientFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					client := mocks.NewMockClient(ctrl)

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						Return(nil).
						AnyTimes().
						Do(func(_ context.Context, username, password string) {
							t.Logf("Auth method called with: %s, %s", username, password)
						})

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						Return(nil).
						AnyTimes().
						Do(func(_ context.Context) {
							t.Logf("UpdatePackages method called")
						})
					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						Return(client, nil).
						AnyTimes()

					client.
						EXPECT().
						EnableService(gomock.Any(), "test").
						Return(nil)
				},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "test" {
					name = "test"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.test", "enabled", "true"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.test",
							Action: tfjson.ActionCreate,
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
					&testutil.StateChecker{
						Addr:     "openwrt_service.test",
						AttrName: "enabled",
						Value:    "true",
					},
				},
			},
		},
	})
}
