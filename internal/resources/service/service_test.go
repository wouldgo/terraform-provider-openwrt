// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
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
						IsEnabled(gomock.Any(), "service#1").
						DoAndReturn(func(_ context.Context, _ string) (bool, error) {
							t.Log("IsEnabled method called")

							return true, nil
						}).
						AnyTimes()

					theRPC.
						EXPECT().
						EnableService(gomock.Any(), "service#1").
						DoAndReturn(func(_ context.Context, _ string) error {
							t.Logf("EnableService method called")
							return nil
						}).
						AnyTimes()

					rpcFactory.
						EXPECT().
						Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any()).
						DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts) (luci.RPC, error) {
							t.Logf("Get method called")
							return theRPC, nil
						}).
						AnyTimes()
				},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "a_service" {
					name = "service#1"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.a_service", "enabled", "true"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
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
						Addr:     "openwrt_service.a_service",
						AttrName: "enabled",
						Value:    "true",
					},
				},
			},
		},
	})
}

func TestAccService_CheckServiceEnable(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			theRPC := mocks.NewMockRPC(ctrl)

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
				IsEnabled(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return true, nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				EnableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("EnableService method called")
					return nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				RestartService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("RestartService method called")
					return nil
				}).
				AnyTimes()

			rpcFactory.
				EXPECT().
				Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any()).
				DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts) (luci.RPC, error) {
					t.Logf("Get method called")
					return theRPC, nil
				}).
				AnyTimes()
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "a_service" {
					name = "service#1"
					enabled = true
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.a_service", "enabled", "true"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
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
						Addr:     "openwrt_service.a_service",
						AttrName: "enabled",
						Value:    "true",
					},
				},
			},

			{
				PreConfig: func() {},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "a_service" {
					name = "service#1"
					enabled = true
					triggers = {
						conf_sha = "123star"
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.a_service", "enabled", "true"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
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
						Addr:     "openwrt_service.a_service",
						AttrName: "enabled",
						Value:    "true",
					},
				},
			},
		},
	})
}

func TestAccService_CheckServiceDisable(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			theRPC := mocks.NewMockRPC(ctrl)

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
				IsEnabled(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return false, nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				DisableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("DisableService method called")
					return nil
				}).
				AnyTimes()

			rpcFactory.
				EXPECT().
				Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any()).
				DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts) (luci.RPC, error) {
					t.Logf("Get method called")
					return theRPC, nil
				}).
				AnyTimes()
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "a_service" {
					name = "service#1"
					enabled = false
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.a_service", "enabled", "false"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
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
						Addr:     "openwrt_service.a_service",
						AttrName: "enabled",
						Value:    "false",
					},
				},
			},
		},
	})
}

func TestAccService_CheckServiceEnableDisable(t *testing.T) {
	os.Setenv("TF_ACC", "1")    //nolint:errcheck
	defer os.Unsetenv("TF_ACC") //nolint:errcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rpcFactory := mocks.NewMockRPCFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(rpcFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			theRPC := mocks.NewMockRPC(ctrl)

			theRPC.
				EXPECT().
				UpdatePackages(gomock.Any()).
				DoAndReturn(func(_ context.Context) error {
					t.Logf("UpdatePackages method called")
					return nil
				}).
				AnyTimes()

			isEnableToTrueCalled := theRPC.
				EXPECT().
				IsEnabled(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return true, nil
				}).
				Times(2)

			theRPC.
				EXPECT().
				IsEnabled(gomock.Any(), "service#1").
				After(isEnableToTrueCalled).
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return false, nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				EnableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("EnableService method called")
					return nil
				}).
				AnyTimes()

			theRPC.
				EXPECT().
				DisableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("DisableService method called")
					return nil
				}).
				AnyTimes()

			rpcFactory.
				EXPECT().
				Get(gomock.Any(), "http://test.lan:8080", "root", "test", gomock.Any()).
				DoAndReturn(func(_ context.Context, _, _, _ string, _ luci.Timeouts) (luci.RPC, error) {
					t.Logf("Get method called")
					return theRPC, nil
				}).
				AnyTimes()
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "a_service" {
					name = "service#1"
					enabled = true
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.a_service", "enabled", "true"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
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
						Addr:     "openwrt_service.a_service",
						AttrName: "enabled",
						Value:    "true",
					},
				},
			},

			{
				PreConfig: func() {},
				Config: `
				provider "openwrt" {
					user     = "root"
					password = "test"
					remote   = "http://test.lan:8080"
				}

				resource "openwrt_service" "a_service" {
					name = "service#1"
					enabled = false
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_service.a_service", "enabled", "false"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
							Action: tfjson.ActionUpdate,
						},
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						&testutil.ActionPlanChecker{
							Addr:   "openwrt_service.a_service",
							Action: tfjson.ActionNoop,
						},
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					&testutil.StateChecker{
						Addr:     "openwrt_service.a_service",
						AttrName: "enabled",
						Value:    "false",
					},
				},
			},
		},
	})
}
