// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package opkg

import (
	"context"
	"fmt"
	"strings"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	OpkgTimeoutSchemaAttribute = schema.SingleNestedAttribute{
		MarkdownDescription: `Opkg operations timeout configuration`,
		Description:         `Opkg operations timeout configuration`,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"update_packages": schema.StringAttribute{
				MarkdownDescription: `Update packages RPC timeout value`,
				Description:         `Update packages RPC timeout value`,
				Optional:            true,
			},
			"check_package": schema.StringAttribute{
				MarkdownDescription: `Check package RPC timeout value`,
				Description:         `Check package RPC timeout value`,
				Optional:            true,
			},
			"install_packages": schema.StringAttribute{
				MarkdownDescription: `Install packages RPC timeout value`,
				Description:         `Install packages RPC timeout value`,
				Optional:            true,
			},
			"remove_packages": schema.StringAttribute{
				MarkdownDescription: `Remove packages RPC timeout value`,
				Description:         `Remove packages RPC timeout value`,
				Optional:            true,
			},
		},
	}
)

type OpkgTimeoutsModel struct {
	UpdatePackagesTimeout  types.String `tfsdk:"update_packages"`
	CheckPackageTimeout    types.String `tfsdk:"check_package"`
	InstallPackagesTimeout types.String `tfsdk:"install_packages"`
	RemovePackagesTimeout  types.String `tfsdk:"remove_packages"`
}

type opkgModel struct {
	Packages types.List `tfsdk:"packages"`
}

type opkgResource struct {
	opkgFacade rpc.OpkgFacade
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

func (c *opkgResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := req.ProviderData
	if data == nil {
		return
	}
	opkgFacade, ok := data.(rpc.OpkgFacade)
	if !ok {
		resp.Diagnostics.AddError("failed to get opkg facade", "")
		return
	}

	err := opkgFacade.UpdatePackages(ctx)
	if err != nil {
		resp.Diagnostics.AddError("packages update in error", err.Error())
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

	toInstall := make([]string, 0, len(plan.Packages.Elements()))
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
			toInstall = append(toInstall, valueStr)
		}
	}

	if len(toInstall) > 0 {

		if err := c.opkgFacade.InstallPackages(ctx, toInstall...); err != nil {
			resp.Diagnostics.AddError("failed to install packages", fmt.Sprintf("[%s]: %v", strings.Join(toInstall, ", "), err))
			return
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
	packagesToInstall := make([]string, 0, len(planSet))
	for aPackageInPlan := range planSet {
		if _, aPackageInPlanAlsoInState := stateSet[aPackageInPlan]; !aPackageInPlanAlsoInState { // new package
			packagesToInstall = append(packagesToInstall, aPackageInPlan)
		} else { // already existing do nothing
			resp.Diagnostics.AddWarning("package already installed", aPackageInPlan)
		}
	}

	if len(packagesToInstall) > 0 {
		if err := c.opkgFacade.InstallPackages(ctx, packagesToInstall...); err != nil {
			resp.Diagnostics.AddError("failed to install package", fmt.Sprintf("[%s]: %v", strings.Join(packagesToInstall, ", "), err))
			return
		}
	}

	// removals
	packagesToRemove := make([]string, 0, len(stateSet))
	for aPackageInState := range stateSet {
		if _, aPackageInStateAlsoInPlan := planSet[aPackageInState]; !aPackageInStateAlsoInPlan { // package no more in plan
			packagesToRemove = append(packagesToRemove, aPackageInState)
		}
	}

	if len(packagesToRemove) > 0 {
		if err := c.opkgFacade.RemovePackages(ctx, packagesToRemove...); err != nil {
			resp.Diagnostics.AddError("failed to remove package", fmt.Sprintf("[%s]: %v", strings.Join(packagesToRemove, ", "), err))
			return
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
