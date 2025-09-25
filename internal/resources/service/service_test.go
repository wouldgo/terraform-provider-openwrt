// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
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

					client.
						EXPECT().
						IsEnabled(gomock.Any(), "service#1").
						DoAndReturn(func(_ context.Context, _ string) (bool, error) {
							t.Log("IsEnabled method called")

							return true, nil
						}).
						AnyTimes()

					client.
						EXPECT().
						EnableService(gomock.Any(), "service#1").
						DoAndReturn(func(_ context.Context, _ string) error {
							t.Logf("EnableService method called")
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

	clientFactory := mocks.NewMockClientFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(clientFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
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

			client.
				EXPECT().
				IsEnabled(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return true, nil
				}).
				AnyTimes()

			client.
				EXPECT().
				EnableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("EnableService method called")
					return nil
				}).
				AnyTimes()

			client.
				EXPECT().
				RestartService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("RestartService method called")
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

	clientFactory := mocks.NewMockClientFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(clientFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
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

			client.
				EXPECT().
				IsEnabled(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return false, nil
				}).
				AnyTimes()

			client.
				EXPECT().
				DisableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("DisableService method called")
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

	clientFactory := mocks.NewMockClientFactory(ctrl)
	testAccProtoV6ProviderFactories := testutil.TestAccFactories(clientFactory)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
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

			isEnableToTrueCalled := client.
				EXPECT().
				IsEnabled(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return true, nil
				}).
				Times(2)

			client.
				EXPECT().
				IsEnabled(gomock.Any(), "service#1").
				After(isEnableToTrueCalled).
				DoAndReturn(func(_ context.Context, _ string) (bool, error) {
					t.Log("IsEnabled method called")

					return false, nil
				}).
				AnyTimes()

			client.
				EXPECT().
				EnableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("EnableService method called")
					return nil
				}).
				AnyTimes()

			client.
				EXPECT().
				DisableService(gomock.Any(), "service#1").
				DoAndReturn(func(_ context.Context, _ string) error {
					t.Log("DisableService method called")
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
