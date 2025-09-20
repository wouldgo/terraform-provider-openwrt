// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package fs

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fileModel struct {
	Path    types.String `tfsdk:"path"`
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
}

type fileResource struct {
	provider api.Client
}

func NewFileResource() resource.Resource {
	return &fileResource{}
}

func (c fileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_file", req.ProviderTypeName)
}

func (c fileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Write configuration files to an arbitray path on the OpenWRT router.",
		Description:         "Write configuration files to an arbitray path on the OpenWRT router.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Path where file has to be.",
				Description:         "Path where file has to be.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the file.",
				Description:         "Name of the file.",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The content of the file.",
				Description:         "The content of the file.",
				Required:            true,
			},
		},
	}
}

func (c *fileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (c fileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := path.Join(plan.Path.ValueString(), plan.Name.ValueString())

	if err := c.provider.Writefile(ctx, path, []byte(plan.Content.ValueString())); err != nil {
		resp.Diagnostics.AddError("Failed to write file", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (c fileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := path.Join(state.Path.ValueString(), state.Name.ValueString())
	b, err := c.provider.ReadFile(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read config %q", state.Name.ValueString()), err.Error())
		return
	}

	// Logic taken from the local_file provider
	if !bytes.Equal(b, []byte(state.Content.ValueString())) {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (c fileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fileModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := path.Join(plan.Path.ValueString(), plan.Name.ValueString())
	if err := c.provider.Writefile(ctx, path, []byte(plan.Content.ValueString())); err != nil {
		resp.Diagnostics.AddError("Failed to write file", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (c fileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fileModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := path.Join(state.Path.ValueString(), state.Name.ValueString())
	err := c.provider.RemoveFile(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete config file %q", state.Name.ValueString()), err.Error())
		return
	}
}
