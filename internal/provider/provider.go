// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/system"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OpenWRTProvider
type OpenWRTProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpenWRTProvider{
			version: version,
		}
	}
}

type OpenWRTProviderModel struct {
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Remote   types.String `tfsdk:"remote"`
}

func (p *OpenWRTProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openwrt"
	resp.Version = p.version
}

func (p *OpenWRTProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"remote": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Required:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Required:            true,
			},
		},
	}
}

func (p *OpenWRTProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OpenWRTProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c := api.NewClient(data.Remote.ValueString())
	if err := c.Auth(data.User.ValueString(), data.Password.ValueString()); err != nil {
		panic(err)
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *OpenWRTProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		system.NewSystemResource,
	}
}

func (p *OpenWRTProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *OpenWRTProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *OpenWRTProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}
