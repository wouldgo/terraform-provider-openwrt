package provider

import (
	"context"
	"os"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/fs"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/opkg"
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

	c, err = p.clientFactory.Get(remoteUrl)
	if err != nil {
		resp.Diagnostics.AddError("failed to instantiate remote client from env variable", err.Error())
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
