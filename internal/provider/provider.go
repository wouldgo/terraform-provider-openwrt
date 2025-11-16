// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/fs"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/opkg"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/service"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/system"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = (*OpenWRTProvider)(nil)

	openWRTRemoteEnv,
	openWRTRemoteEnvSet = os.LookupEnv("OPENWRT_REMOTE")

	openWRTUserEnv,
	openWRTUserEnvSet = os.LookupEnv("OPENWRT_USER")

	openWRTPasswordEnv,
	openWRTPasswordEnvSet = os.LookupEnv("OPENWRT_PASSWORD")
)

// OpenWRTProvider
type OpenWRTProvider struct {
	version       string
	clientFactory api.ClientFactory
}

func New(version string, clientFactory api.ClientFactory) func() provider.Provider {
	return func() provider.Provider {
		return &OpenWRTProvider{
			version:       version,
			clientFactory: clientFactory,
		}
	}
}

type OpenWRTProviderModel struct {
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Remote   types.String `tfsdk:"remote"`

	ApiTimeouts *api.TimeoutsModel `tfsdk:"api_timeouts"`
}

func (p *OpenWRTProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openwrt"
	resp.Version = p.version
}

func (p *OpenWRTProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This provider connets to openwrt routers through the UCI JSON RPC API.

The JSON RPC API requires a couple of packages to be used. Please see [Using the JSON-RPC API](https://github.com/openwrt/luci/blob/master/docs/JsonRpcHowTo.md) from openwrt.`,
		Description: "Terraform, or OpenTofu, provider to manage openwrt routers",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				MarkdownDescription: `The password of the account. Optionally OPENWRT_USER env variable can be set and used to specify the user. One between this attribute or the env variable must be set`,
				Description:         `The password of the account. Optionally OPENWRT_USER env variable can be set and used to specify the user. One between this attribute or the env variable must be set`,
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: `The URL of the JSON RPC API. Optionally OPENWRT_PASSWORD env variable can be set and used to specify the password. One between this attribute or the env variable must be set`,
				Description:         `The URL of the JSON RPC API. Optionally OPENWRT_PASSWORD env variable can be set and used to specify the password. One between this attribute or the env variable must be set`,
				Optional:            true,
			},
			"remote": schema.StringAttribute{
				MarkdownDescription: `The username of the admin account. Optionally OPENWRT_REMOTE env variable can be set and used to specify the remote url. One between this attribute or the env variable must be set`,
				Description:         `The username of the admin account. Optionally OPENWRT_REMOTE env variable can be set and used to specify the remote url. One between this attribute or the env variable must be set`,
				Optional:            true,
			},
			"api_timeouts": api.TimeoutSchemaAttribute,
		},
	}
}

func (p *OpenWRTProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var (
		data OpenWRTProviderModel
		c    api.Client
		err  error
	)

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remoteUrl := data.Remote.ValueString()
	if openWRTRemoteEnvSet {
		remoteUrl = openWRTRemoteEnv
	}

	apiTimeouts, err := p.clientFactory.ParseTimeouts(ctx, data.ApiTimeouts)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse timeouts", err.Error())
		return
	}
	c, err = p.clientFactory.Get(ctx, remoteUrl, apiTimeouts)
	if err != nil {
		resp.Diagnostics.AddError("failed to instantiate remote client", err.Error())
		return
	}

	username, password := data.User.ValueString(), data.Password.ValueString()
	if openWRTUserEnvSet && openWRTPasswordEnvSet {
		username = openWRTUserEnv
		password = openWRTPasswordEnv
	}
	err = c.Auth(ctx, username, password)
	if err != nil {
		resp.Diagnostics.AddError("failed to auth towards openwrt API", err.Error())
		return
	}

	err = c.UpdatePackages(ctx)
	if err != nil {
		resp.Diagnostics.AddError("packages update in error", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *OpenWRTProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		system.NewSystemResource,
		fs.NewConfigFileResource,
		fs.NewFileResource,
		opkg.NewOpkgResource,
		service.NewServiceResource,
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
