// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package opkg_test

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/testutil"

	"github.com/foxboron/terraform-provider-openwrt/mocks"
	"go.uber.org/mock/gomock"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func prepareMockProvider(
	t *testing.T,
	rpcFactory *mocks.MockRPCFactory,
	theRPC *mocks.MockRPC,
) {
	rpcFactory.
		EXPECT().
		Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts, _ luci.Attempts) (luci.RPC, error) {
			t.Logf("Get method called")
			return theRPC, nil
		}).
		AnyTimes()
}

func TestAccOpkg_AllDepsAreMissing(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					luciRPC := mocks.NewMockRPC(ctrl)

					prepareMockProvider(t, rpcFactory, luciRPC)

					luciRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					checkPackagesNotInstalled := luciRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return luci.PackageInfo{
								Version: "",
								Status: luci.Status{
									Installed: false,
								},
							}, nil
						}).
						Times(1)

					luciRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return luci.PackageInfo{
								Version: "test",
								Status: luci.Status{
									Installed: true,
								},
							}, nil
						}).
						AnyTimes().
						After(checkPackagesNotInstalled)

					luciRPC.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						DoAndReturn(func(_ context.Context, _ ...string) error {
							t.Logf("InstallPackages method called")
							return nil
						}).
						Times(1)

					//Teardown resource
					luciRPC.
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

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					theRPC := mocks.NewMockRPC(ctrl)

					prepareMockProvider(t, rpcFactory, theRPC)

					theRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					theRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return luci.PackageInfo{
								Version: "test",
								Status: luci.Status{
									Installed: true,
								},
							}, nil
						}).
						AnyTimes()

					//Teardown resource
					theRPC.
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

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
			theRPC := mocks.NewMockRPC(ctrl)

			prepareMockProvider(t, rpcFactory, theRPC)

			theRPC.
				EXPECT().
				UpdatePackages(gomock.Any()).
				DoAndReturn(func(_ context.Context) error {
					t.Logf("UpdatePackages method called")
					return nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				CheckPackage(gomock.Any(), "curl").
				DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
					t.Logf("CheckPackage method called")
					return luci.PackageInfo{
						Version: "test",
						Status: luci.Status{
							Installed: true,
						},
					}, nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				InstallPackages(gomock.Any(), "wget").
				DoAndReturn(func(ctx context.Context, s ...string) error {
					t.Logf("InstallPackages method called")
					return nil
				}).
				Times(1)

			theRPC.
				EXPECT().
				CheckPackage(gomock.Any(), "wget").
				DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
					t.Logf("CheckPackage method called")
					return luci.PackageInfo{
						Version: "test",
						Status: luci.Status{
							Installed: true,
						},
					}, nil
				}).
				AnyTimes()

			//Teardown resource
			theRPC.
				EXPECT().
				RemovePackages(gomock.Any(), "curl").
				DoAndReturn(func(ctx context.Context, s ...string) error {
					t.Logf("RemovePackages method called")
					return nil
				}).
				Times(1)

			theRPC.
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

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					rpcFactory.
						EXPECT().
						Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts, _ luci.Attempts) (luci.RPC, error) {
							t.Logf("Get method called")
							return nil, luci.ErrMissingRemoteBaseURL
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
				ExpectError: regexp.MustCompile(luci.ErrMissingRemoteBaseURL.Error()),
			},

			{
				PreConfig: func() {
					theRPC := mocks.NewMockRPC(ctrl)

					rpcFactory.
						EXPECT().
						Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts, _ luci.Attempts) (luci.RPC, error) {
							t.Logf("Get method called")
							return theRPC, nil
						}).
						Times(1)

					theRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(ctx context.Context) error {
							t.Logf("UpdatePackages method called")
							return api.ErrMarshal
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
					theRPC := mocks.NewMockRPC(ctrl)

					rpcFactory.
						EXPECT().
						Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts, _ luci.Attempts) (luci.RPC, error) {
							t.Logf("Get method called")
							return theRPC, nil
						}).
						Times(1)

					theRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(ctx context.Context) error {
							t.Logf("UpdatePackages method called")
							return luci.ErrFloatExpected
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
				ExpectError: regexp.MustCompile(luci.ErrFloatExpected.Error()),
			},
		},
	})
}

func TestAccOpkg_CheckPackageInCreateIsFailing(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					theRPC := mocks.NewMockRPC(ctrl)

					prepareMockProvider(t, rpcFactory, theRPC)

					theRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					theRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							var zero luci.PackageInfo
							return zero, luci.ErrPackageNotFound
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
				ExpectError: regexp.MustCompile(luci.ErrPackageNotFound.Error()),
			},
		},
	})
}

func TestAccOpkg_InstallPackagesIsFailing(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					theRPC := mocks.NewMockRPC(ctrl)

					prepareMockProvider(t, rpcFactory, theRPC)

					theRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(_ context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					theRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return luci.PackageInfo{
								Version: "",
								Status: luci.Status{
									Installed: false,
								},
							}, nil
						}).
						Times(1)

					theRPC.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s ...string) error {
							t.Logf("InstallPackages method called")
							return luci.ErrFloatExpected
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
				ExpectError: regexp.MustCompile(luci.ErrFloatExpected.Error()),
			},
		},
	})
}

func TestAccOpkg_CheckPackageInUpdateIsFailing(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			// e.g., check environment or emulator availability
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					theRPC := mocks.NewMockRPC(ctrl)

					rpcFactory.
						EXPECT().
						Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts, _ luci.Attempts) (luci.RPC, error) {
							t.Logf("Get method called")
							return theRPC, nil
						}).
						AnyTimes()

					theRPC.
						EXPECT().
						UpdatePackages(gomock.Any()).
						DoAndReturn(func(ctx context.Context) error {
							t.Logf("UpdatePackages method called")
							return nil
						}).
						AnyTimes()

					checkPackagesNotInstalled := theRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							return luci.PackageInfo{
								Version: "",
								Status: luci.Status{
									Installed: false,
								},
							}, nil
						}).
						Times(1)

					theRPC.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s string) (luci.PackageInfo, error) {
							t.Logf("CheckPackage method called")
							var zero luci.PackageInfo
							return zero, luci.ErrPackageNotFound
						}).
						AnyTimes().
						After(checkPackagesNotInstalled)

					theRPC.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						DoAndReturn(func(ctx context.Context, s ...string) error {
							t.Logf("InstallPackages method called")
							return nil
						}).
						Times(1)

					//Teardown resource
					theRPC.
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
				ExpectError: regexp.MustCompile(luci.ErrPackageNotFound.Error()),
			},
		},
	})
}
