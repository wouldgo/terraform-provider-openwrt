// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

//go:build test

package testutil

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

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
	var foundActions tfjson.Actions
	for _, aResource := range checkReq.Plan.ResourceChanges {
		if aResource.Address == pC.Addr && slices.Contains(aResource.Change.Actions, pC.Action) {
			raiseErr = false
		}

		if aResource.Address == pC.Addr {
			foundActions = aResource.Change.Actions
		}
	}
	if raiseErr {
		checkResponse.Error = fmt.Errorf("%+v missing %+v action. found %+v", pC.Addr, pC.Action, foundActions)
	}
}

type StateChecker struct {
	Addr     string
	AttrName string
	Value    string
}

func (cS *StateChecker) CheckState(_ context.Context,
	checkReq statecheck.CheckStateRequest, checkResponse *statecheck.CheckStateResponse) {
	for _, aResource := range checkReq.State.Values.RootModule.Resources {
		if aResource.Address == cS.Addr {
			for attributeName, attributeValue := range aResource.AttributeValues {
				if attributeName != cS.AttrName {
					continue
				}

				err := cS.verifyAttribute(attributeValue)
				if err != nil {
					checkResponse.Error = fmt.Errorf("state is incosistent: %s is %+v instead of %+v: %w",
						attributeName,
						attributeValue,
						cS.Value,
						err,
					)
					return
				}
			}
		}
	}
}

func (cS *StateChecker) verifyAttribute(attributeValue interface{}) error {
	switch v := attributeValue.(type) {
	case []interface{}:
		var errs []error
		inError := true
		for _, anElm := range v {
			err := cS.verifyAttribute(anElm)
			if err != nil {
				errs = append(errs, err)
			} else {
				inError = false
			}
		}
		if inError {
			return errors.Join(errs...)
		}
		return nil
	case string:
		if v == cS.Value {
			return nil
		}
		return fmt.Errorf("string attribute %+v not equals to %+v", v, cS.Value)
	case bool:
		if strconv.FormatBool(v) == cS.Value {
			return nil
		}
		return fmt.Errorf("boolean attribute %+v not equals to %+v", v, cS.Value)
	default:
		return fmt.Errorf("attribute %+v not managed", attributeValue)
	}
}
