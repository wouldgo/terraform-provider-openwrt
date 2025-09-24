// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package opkg

import (
	"context"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type opkgModel struct {
	Packages types.List `tfsdk:"packages"`
}

type opkgResource struct {
	opkgFacade api.OpkgFacade
}

func NewOpkgResource() resource.Resource {
	return &opkgResource{}
}

func (c opkgResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_opkg", req.ProviderTypeName)
}

func (c opkgResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Install packages on the router",
		Description:         "Install packages on the router",
		Attributes: map[string]schema.Attribute{
			"packages": schema.ListAttribute{
				MarkdownDescription: "The list of packages to install via opkg package manager",
				Description:         "The list of packages to install via opkg package manager",
				ElementType:         types.StringType,
				Required:            true,
			},
		},
	}
}

func (c *opkgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := req.ProviderData
	if data == nil {
		return
	}
	opkgFacade, ok := data.(api.OpkgFacade)
	if !ok {
		resp.Diagnostics.AddError("failed to get opkg facade", "")
		return
	}
	c.opkgFacade = opkgFacade
}

func (c opkgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan opkgModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, aPackage := range plan.Packages.Elements() {
		value, err := aPackage.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError("can not retrieve value", fmt.Sprintf("%s: %v", aPackage.String(), err))
			return
		}
		var valueStr string
		err = value.As(&valueStr)
		if err != nil {
			resp.Diagnostics.AddError("value cannot be read", fmt.Sprintf("value %s not readable as string: %v", value, err))
			return
		}

		re, err := c.opkgFacade.CheckPackage(ctx, valueStr)
		if err != nil {
			resp.Diagnostics.AddError("checking package went in error", fmt.Sprintf("%s: %v", valueStr, err))
			return
		}

		if !re.Status.Installed {
			if err = c.opkgFacade.InstallPackages(ctx, valueStr); err != nil {
				resp.Diagnostics.AddError("failed to install package", fmt.Sprintf("%s: %v", valueStr, err))
				return
			}
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (c opkgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state opkgModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := make([]attr.Value, 0, len(state.Packages.Elements()))
	for _, packageElm := range state.Packages.Elements() {
		packageValue, err := packageElm.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError("can not retrieve value", fmt.Sprintf("%s: %v", packageElm.String(), err))
			return
		}
		var packageValueStr string
		err = packageValue.As(&packageValueStr)
		if err != nil {
			resp.Diagnostics.AddError("value cannot be read", fmt.Sprintf("package value %s not readable as string: %v", packageValue, err))
			return
		}

		re, err := c.opkgFacade.CheckPackage(ctx, packageValueStr)
		if err != nil {
			resp.Diagnostics.AddError("checking package went in error", fmt.Sprintf("%s: %v", packageValueStr, err))
			return
		}

		if re.Status.Installed {
			result = append(result, packageElm)
		}
	}

	state.Packages = basetypes.NewListValueMust(types.StringType, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (c opkgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state opkgModel
	var plan opkgModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planSet := make(map[string]struct{})
	for _, aPackage := range plan.Packages.Elements() {
		value, err := aPackage.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError("can not retrieve value from plan", fmt.Sprintf("%s: %v", aPackage.String(), err))
			return
		}
		var valueStr string
		err = value.As(&valueStr)
		if err != nil {
			resp.Diagnostics.AddError("value cannot be read", fmt.Sprintf("package value %s not readable as string: %v", value, err))
			return
		}

		planSet[valueStr] = struct{}{}
	}

	stateSet := make(map[string]struct{})
	for _, aPackage := range state.Packages.Elements() {
		value, err := aPackage.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError("can not retrieve value from state", fmt.Sprintf("%s: %v", aPackage.String(), err))
			return
		}
		var valueStr string
		err = value.As(&valueStr)
		if err != nil {
			resp.Diagnostics.AddError("value cannot be read", fmt.Sprintf("package value %s not readable as string: %v", value, err))
			return
		}

		stateSet[valueStr] = struct{}{}
	}

	// additions
	for aPackageInPlan := range planSet {
		if _, aPackageInPlanAlsoInState := stateSet[aPackageInPlan]; !aPackageInPlanAlsoInState { // new package
			if err := c.opkgFacade.InstallPackages(ctx, aPackageInPlan); err != nil {
				resp.Diagnostics.AddError("failed to install package", fmt.Sprintf("%s: %v", aPackageInPlan, err))
				return
			}
		} else { // already existing do nothing
			resp.Diagnostics.AddWarning("package already installed", aPackageInPlan)
		}
	}

	// removals
	for aPackageInState := range stateSet {
		if _, aPackageInStateAlsoInPlan := planSet[aPackageInState]; !aPackageInStateAlsoInPlan { // package no more in plan
			if err := c.opkgFacade.RemovePackages(ctx, aPackageInState); err != nil {
				resp.Diagnostics.AddError("failed to remove package", fmt.Sprintf("%s: %v", aPackageInState, err))
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (c opkgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state opkgModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, aPackage := range state.Packages.Elements() {
		value, err := aPackage.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError("can not retrieve value", fmt.Sprintf("%s: %v", aPackage.String(), err))
			return
		}
		var valueStr string
		err = value.As(&valueStr)
		if err != nil {
			resp.Diagnostics.AddError("value cannot be read", fmt.Sprintf("package value %s not readable as string: %v", value, err))
			return
		}

		err = c.opkgFacade.RemovePackages(ctx, valueStr)
		if err != nil {
			resp.Diagnostics.AddError("removing package went in error", fmt.Sprintf("%s: %v", valueStr, err))
			return
		}
	}
}
