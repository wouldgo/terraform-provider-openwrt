// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
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
	version    string
	rpcFactory rpc.RPCFactory
}

func New(version string, rpcFactory rpc.RPCFactory) func() provider.Provider {
	return func() provider.Provider {
		return &OpenWRTProvider{
			version:    version,
			rpcFactory: rpcFactory,
		}
	}
}

type OpenWRTProviderModel struct {
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Remote   types.String `tfsdk:"remote"`

	APITimeouts *TimeoutsModel `tfsdk:"api_timeouts"`
	APIAttempts *AttemptsModel `tfsdk:"api_attempts"`
}

func (p *OpenWRTProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openwrt"
	resp.Version = p.version
}

func (p *OpenWRTProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This provider connets to openwrt routers through the UCI JSON RPC API.

The JSON RPC API requires a couple of packages to be used. Please see [Using the JSON-RPC API](https://htmlpreview.github.io/?https://raw.githubusercontent.com/openwrt/luci/master/docs/api/index.html) from openwrt.`,
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
			"api_timeouts": schema.SingleNestedAttribute{
				MarkdownDescription: "Timeout configuration for the specific RPC calls. The main purpose of this optional configuration is to fine tune the default timeouts for longer API interaction (e.g. update packages, list packages, ...)",
				Description:         "Timeout configuration for the specific RPC calls",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"auth": schema.StringAttribute{
						MarkdownDescription: `Authentication RPC timeout value`,
						Description:         `Authentication RPC timeout value`,
						Optional:            true,
					},
					"fs":      fs.FsTimeoutSchemaAttribute,
					"opkg":    opkg.OpkgTimeoutSchemaAttribute,
					"service": service.ServiceTimeoutSchemaAttribute,
					"uci":     system.UciTimeoutSchemaAttribute,
				},
			},
			"api_attempts": schema.SingleNestedAttribute{
				MarkdownDescription: "Attempts configuration for each RPC calls. The main purpose of this optional configuration is to overcome the timeout of the luci API because of the limited resources (e.g. during package installation) or the restard of luci because of related packages installed via the provider",
				Description:         "Attempts configuration for each RPC calls.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"fs":      fs.FsAttemptsSchemaAttribute,
					"opkg":    opkg.OpkgAttemptsSchemaAttribute,
					"service": service.ServiceAttemptsSchemaAttribute,
					"uci":     system.UciAttemptsSchemaAttribute,
				},
			},
		},
	}
}

func (p *OpenWRTProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var (
		data OpenWRTProviderModel
		r    rpc.RPC
		err  error
	)

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	remoteURL := data.Remote.ValueString()
	if openWRTRemoteEnvSet {
		remoteURL = openWRTRemoteEnv
	}

	apiTimeouts, err := parseTimeouts(ctx, data.APITimeouts)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse timeouts", err.Error())
		return
	}

	apiAttempts, err := parseAttempts(ctx, data.APIAttempts)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse attempts", err.Error())
		return
	}

	username, password := data.User.ValueString(), data.Password.ValueString()
	if openWRTUserEnvSet && openWRTPasswordEnvSet {
		username = openWRTUserEnv
		password = openWRTPasswordEnv
	}

	r, err = p.rpcFactory.Get(
		ctx,
		remoteURL,
		username,
		password,
		apiTimeouts,
		apiAttempts,
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to instantiate remote client", err.Error())
		return
	}

	resp.DataSourceData = r
	resp.ResourceData = r
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
