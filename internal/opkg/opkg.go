package opkg

import (
	"context"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OpkgModel struct {
	Packages types.List `tfsdk:"packages"`
}

type OpkgResource struct {
	provider api.Client
}

func NewOpkgResource() resource.Resource {
	return &OpkgResource{}
}

func (c OpkgResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_opkg", req.ProviderTypeName)
}

func (c OpkgResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"packages": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (c *OpkgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := req.ProviderData
	if data == nil {
		return
	}
	provider, ok := data.(api.Client)
	if !ok {
		resp.Diagnostics.AddError("Failed to get api client", "")
		return
	}
	c.provider = provider
}

func (c OpkgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OpkgModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//TODO install packages
	for _, v := range plan.Packages.Elements() {
		value, err := v.ToTerraformValue(ctx)
		if err != nil {
			resp.Diagnostics.AddError("can not retrieve value", v.String())
			return
		}
		var valueStr string
		value.As(&valueStr)

		re, err := c.provider.CheckPackage(valueStr)
		if err != nil {
			resp.Diagnostics.AddError("checking package went in error", valueStr)
		}

		fmt.Println(re)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (c OpkgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OpkgModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	test := state.Packages.Elements()
	fmt.Println(test)
	//TODO verify packages presence
	// for _, value := range state.Packages.Elements() {

	// 	c.provider.CheckPackage()
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (c OpkgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OpkgModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//TODO update packages
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (c OpkgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OpkgModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//TODO delete packages
}
