package provider_test

import (
	"context"
	"fmt"
	"slices"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

var (
	_ plancheck.PlanCheck   = (*actionPlanChecker)(nil)
	_ statecheck.StateCheck = (*stateChecker)(nil)
)

type actionPlanChecker struct {
	addr   string
	action tfjson.Action
}

func (pC *actionPlanChecker) CheckPlan(_ context.Context, checkReq plancheck.CheckPlanRequest, checkResponse *plancheck.CheckPlanResponse) {
	raiseErr := true
	for _, aResource := range checkReq.Plan.ResourceChanges {
		if aResource.Address == pC.addr && slices.Contains(aResource.Change.Actions, pC.action) {
			raiseErr = false
		}
	}
	if raiseErr {
		checkResponse.Error = fmt.Errorf("missing %+v action", pC.action)
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
