// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package opkg_test

import (
	"context"
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/testutil"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/mocks"
	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccOpkg_AllDepsAreMissing(t *testing.T) {
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
						DoAndReturn(func(_ context.Context, username, password string) error {
							t.Logf("Auth method called with: %s, %s", username, password)
							return nil
						}).
						AnyTimes()

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(_ string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						AnyTimes()

					checkPackagesNotInstalled := client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return &api.PackageInfo{
								Version: "",
								Status: api.Status{
									Installed: false,
								},
							}, nil
						}).
						Times(1)

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return &api.PackageInfo{
								Version: "test",
								Status: api.Status{
									Installed: true,
								},
							}, nil
						}).
						AnyTimes().
						After(checkPackagesNotInstalled)

					client.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ ...string) error {
							t.Logf("InstallPackages method called")
							return nil
						}).
						Times(1)

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ ...string) error {
							t.Logf("RemovePackages method called")
							return nil
						}).
						Times(1)
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_opkg.test",
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
						Addr:     "openwrt_opkg.test",
						AttrName: "packages",
						Value:    "curl",
					},
				},
			},
		},
	})
}

func TestAccOpkg_NoDepsAreMissing(t *testing.T) {
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
						DoAndReturn(func(_ context.Context, username, password string) error {
							t.Logf("Auth method called with: %s, %s", username, password)
							return nil
						}).
						AnyTimes()

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						AnyTimes()

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return &api.PackageInfo{
								Version: "test",
								Status: api.Status{
									Installed: true,
								},
							}, nil
						}).
						AnyTimes()

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s ...string) error {
							t.Logf("RemovePackages method called")
							return nil
						}).
						Times(1)
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_opkg.test",
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
						Addr:     "openwrt_opkg.test",
						AttrName: "packages",
						Value:    "curl",
					},
				},
			},
		},
	})
}

func TestAccOpkg_OneDepencyIsMissing(t *testing.T) {
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
			client := mocks.NewMockClient(ctrl)

			client.
				EXPECT().
				Auth(gomock.Any(), "root", "test").
				DoAndReturn(func(_ context.Context, username, password string) error {
					t.Logf("Auth method called with: %s, %s", username, password)
					return nil
				}).
				AnyTimes()

			client.
				EXPECT().
				UpdatePackages(gomock.Any()).
				DoAndReturn(func(ctx context.Context) error {
					t.Logf("UpdatePackages method called")
					return nil
				}).
				AnyTimes()

			clientFactory.
				EXPECT().
				Get("http://test.lan:8080").
				DoAndReturn(func(remote string) (api.Client, error) {
					t.Logf("Get method called with: %s", remote)
					return client, nil
				}).
				AnyTimes()

			client.
				EXPECT().
				CheckPackage(gomock.Any(), "curl").
				DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
					t.Logf("CheckPackage method called")
					return &api.PackageInfo{
						Version: "test",
						Status: api.Status{
							Installed: true,
						},
					}, nil
				}).
				AnyTimes()

			client.
				EXPECT().
				InstallPackages(gomock.Any(), "wget").
				DoAndReturn(func(ctx context.Context, s ...string) error {
					t.Logf("InstallPackages method called")
					return nil
				}).
				Times(1)

			client.
				EXPECT().
				CheckPackage(gomock.Any(), "wget").
				DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
					t.Logf("CheckPackage method called")
					return &api.PackageInfo{
						Version: "test",
						Status: api.Status{
							Installed: true,
						},
					}, nil
				}).
				AnyTimes()

			//Teardown resource
			client.
				EXPECT().
				RemovePackages(gomock.Any(), "curl").
				DoAndReturn(func(ctx context.Context, s ...string) error {
					t.Logf("RemovePackages method called")
					return nil
				}).
				Times(1)

			client.
				EXPECT().
				RemovePackages(gomock.Any(), "wget").
				DoAndReturn(func(ctx context.Context, s ...string) error {
					t.Logf("RemovePackages method called")
					return nil
				}).
				Times(1)
		},
		Steps: []resource.TestStep{
			{
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_opkg.test",
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
						Addr:     "openwrt_opkg.test",
						AttrName: "packages",
						Value:    "curl",
					},
				},
			},
			{
				Config: `
				provider "openwrt" {
					 user = "root"
						password = "test"
						remote = "http://test.lan:8080"
				}
        resource "openwrt_opkg" "test" {
          packages = ["curl", "wget"]
        }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_opkg.test", "packages.0", "curl"),
					resource.TestCheckResourceAttr("openwrt_opkg.test", "packages.1", "wget"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_opkg.test",
							Action: tfjson.ActionUpdate,
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
						Addr:     "openwrt_opkg.test",
						AttrName: "packages",
						Value:    "curl",
					},
				},
			},
			{
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_opkg.test",
							Action: tfjson.ActionUpdate,
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
						Addr:     "openwrt_opkg.test",
						AttrName: "packages",
						Value:    "curl",
					},
				},
			},
		},
	})
}

func TestAcc_ProviderApiAreFailing(t *testing.T) {
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
					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return nil, api.ErrMissingUrl
						}).
						Times(1)
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
				ExpectError: regexp.MustCompile(api.ErrMissingUrl.Error()),
			},

			{
				PreConfig: func() {
					client := mocks.NewMockClient(ctrl)
					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						Times(1)

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						DoAndReturn(func(ctx context.Context, s1, s2 string) error {
							t.Logf("Auth method called")
							return errors.Join(api.ErrMarshal, errors.New("mon petit json"))
						}).
						Times(1)
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
				ExpectError: regexp.MustCompile(api.ErrMarshal.Error()),
			},

			{
				PreConfig: func() {
					client := mocks.NewMockClient(ctrl)
					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						Times(1)

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						DoAndReturn(func(ctx context.Context, s1, s2 string) error {
							t.Logf("Auth method called")
							return nil
						}).
						Times(1)

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(ctx context.Context) error {
							t.Logf("UpdatePackages method called")
							return api.ErrFloatExpected
						}).
						Times(1)
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
				ExpectError: regexp.MustCompile(api.ErrFloatExpected.Error()),
			},
		},
	})
}

func TestAccOpkg_CheckPackageInCreateIsFailing(t *testing.T) {
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

					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						AnyTimes()

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						DoAndReturn(func(_ context.Context, username, password string) error {
							t.Logf("Auth method called with: %s, %s", username, password)
							return nil
						}).
						Times(2)

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						Times(2)

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return nil, api.ErrPackageNotFound
						}).
						Times(1)
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
				ExpectError: regexp.MustCompile(api.ErrPackageNotFound.Error()),
			},
		},
	})
}

func TestAccOpkg_InstallPackagesIsFailing(t *testing.T) {
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

					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						AnyTimes()

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						DoAndReturn(func(ctx context.Context, s1, s2 string) error {
							t.Logf("Auth method called")
							return nil
						}).
						Times(2)

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(ctx context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						Times(2)

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return &api.PackageInfo{
								Version: "",
								Status: api.Status{
									Installed: false,
								},
							}, nil
						}).
						Times(1)

					client.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s ...string) error {
							t.Logf("InstallPackages method called")
							return api.ErrFloatExpected
						}).
						Times(1)
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
				ExpectError: regexp.MustCompile(api.ErrFloatExpected.Error()),
			},
		},
	})
}

func TestAccOpkg_CheckPackageInUpdateIsFailing(t *testing.T) {
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

					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						DoAndReturn(func(s string) (api.Client, error) {
							t.Logf("Get method called")
							return client, nil
						}).
						AnyTimes()

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						Return(nil).
						DoAndReturn(func(ctx context.Context, s1, s2 string) error {
							t.Logf("Auth method called")
							return nil
						}).
						AnyTimes()

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(ctx context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					checkPackagesNotInstalled := client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return &api.PackageInfo{
								Version: "",
								Status: api.Status{
									Installed: false,
								},
							}, nil
						}).
						Times(1)

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (*api.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return nil, api.ErrPackageNotFound
						}).
						AnyTimes().
						After(checkPackagesNotInstalled)

					client.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s ...string) error {
							t.Logf("InstallPackages method called")
							return nil
						}).
						Times(1)

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s ...string) error {
							t.Logf("InstallPackages method called")
							return nil
						}).
						Times(1)
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
				ExpectError: regexp.MustCompile(api.ErrPackageNotFound.Error()),
			},
		},
	})
}
