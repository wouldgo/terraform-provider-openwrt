// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

//go:build test

package testutil

import (
	"context"
	"fmt"
	"slices"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/provider"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

var (
	_ plancheck.PlanCheck   = (*ActionPlanChecker)(nil)
	_ statecheck.StateCheck = (*StateChecker)(nil)
)

// func testAccPreCheck(t *testing.T) {
// 	// You can add code here to run prior to any test case execution, for example assertions
// 	// about the appropriate environment variables being set are common to see in a pre-check
// 	// function.
// }

func TestAccFactories(clientFactory api.ClientFactory) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"openwrt": providerserver.NewProtocol6WithError(provider.New("test", clientFactory)()),
	}
}

type ActionPlanChecker struct {
	Addr   string
	Action tfjson.Action
}

func (pC *ActionPlanChecker) CheckPlan(_ context.Context, checkReq plancheck.CheckPlanRequest, checkResponse *plancheck.CheckPlanResponse) {
	raiseErr := true
	for _, aResource := range checkReq.Plan.ResourceChanges {
		if aResource.Address == pC.Addr && slices.Contains(aResource.Change.Actions, pC.Action) {
			raiseErr = false
		}
	}
	if raiseErr {
		checkResponse.Error = fmt.Errorf("missing %+v action", pC.Action)
	}
}

type StateChecker struct {
	Addr string
}

func (cS *StateChecker) CheckState(_ context.Context, checkReq statecheck.CheckStateRequest, checkResponse *statecheck.CheckStateResponse) {
	raiseErr := true
	for _, aResource := range checkReq.State.Values.RootModule.Resources {
		if aResource.Address == cS.Addr {
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
