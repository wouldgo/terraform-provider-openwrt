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
					clientFactory.
						EXPECT().
						Get("http://test.lan:8080").
						Return(client, nil).
						AnyTimes()

					checkPackagesNotInstalled := client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(&api.PackageInfo{
							Version: "",
							Status: api.Status{
								Installed: false,
							},
						}, nil).
						Times(1)

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(&api.PackageInfo{
							Version: "test",
							Status: api.Status{
								Installed: true,
							},
						}, nil).
						AnyTimes().
						After(checkPackagesNotInstalled)

					client.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						Return(nil).
						Times(1)

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						Return(nil).
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
						Addr: "openwrt_opkg.test",
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
						Addr: "openwrt_opkg.test",
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
				AnyTimes().
				Do(func(remote string) {
					t.Logf("Get method called with: %s", remote)
				})

			client.
				EXPECT().
				CheckPackage(gomock.Any(), "curl").
				Return(&api.PackageInfo{
					Version: "test",
					Status: api.Status{
						Installed: true,
					},
				}, nil).
				AnyTimes().
				Do(func(_ context.Context, pack string) {
					t.Logf("CheckPackage called with: %s. Returning that is installed", pack)
				})

			client.
				EXPECT().
				InstallPackages(gomock.Any(), "wget").
				Return(nil).
				Times(1).
				Do(func(_ context.Context, pack string) {
					t.Logf("InstallPackages called with: %s", pack)
				})

			client.
				EXPECT().
				CheckPackage(gomock.Any(), "wget").
				Return(&api.PackageInfo{
					Version: "test",
					Status: api.Status{
						Installed: true,
					},
				}, nil).
				AnyTimes().
				Do(func(_ context.Context, pack string) {
					t.Logf("CheckPackage called with: %s. Returning that is installed", pack)
				})

			//Teardown resource
			client.
				EXPECT().
				RemovePackages(gomock.Any(), "curl").
				Return(nil).
				Times(1).
				Do(func(_ context.Context, pack string) {
					t.Logf("RemovePackages called with: %s", pack)
				})

			client.
				EXPECT().
				RemovePackages(gomock.Any(), "wget").
				Return(nil).
				Times(1).
				Do(func(_ context.Context, pack string) {
					t.Logf("RemovePackages called with: %s", pack)
				})
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
						Addr: "openwrt_opkg.test",
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
						Addr: "openwrt_opkg.test",
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
						Addr: "openwrt_opkg.test",
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
						Return(nil, api.ErrMissingUrl).
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
						Return(client, nil).
						Times(1)

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						Return(errors.Join(api.ErrMarshal, errors.New("mon petit json"))).
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
						Return(client, nil).
						Times(1)

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						Return(nil).
						Times(1)

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						Return(api.ErrFloatExpected).
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
						Return(client, nil).
						AnyTimes()

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						Return(nil).
						Times(2).
						Do(func(_ context.Context, username, password string) {
							t.Logf("Auth method called with: %s, %s", username, password)
						})

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						Return(nil).
						Times(2).
						Do(func(_ context.Context) {
							t.Logf("UpdatePackages method called")
						})

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(nil, api.ErrPackageNotFound).
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
						Return(client, nil).
						AnyTimes()

					client.
						EXPECT().
						Auth(gomock.Any(), "root", "test").
						Return(nil).
						Times(2).
						Do(func(_ context.Context, username, password string) {
							t.Logf("Auth method called with: %s, %s", username, password)
						})

					client.
						EXPECT().
						UpdatePackages(gomock.Any()).
						Return(nil).
						Times(2).
						Do(func(_ context.Context) {
							t.Logf("UpdatePackages method called")
						})

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(&api.PackageInfo{
							Version: "",
							Status: api.Status{
								Installed: false,
							},
						}, nil).
						Times(1)

					client.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						Return(api.ErrFloatExpected).
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
						Return(client, nil).
						AnyTimes()

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

					checkPackagesNotInstalled := client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(&api.PackageInfo{
							Version: "",
							Status: api.Status{
								Installed: false,
							},
						}, nil).
						Times(1)

					client.
						EXPECT().
						CheckPackage(gomock.Any(), "curl").
						Return(nil, api.ErrPackageNotFound).
						AnyTimes().
						After(checkPackagesNotInstalled)

					client.
						EXPECT().
						InstallPackages(gomock.Any(), "curl").
						Return(nil).
						Times(1)

					//Teardown resource
					client.
						EXPECT().
						RemovePackages(gomock.Any(), "curl").
						Return(nil).
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
