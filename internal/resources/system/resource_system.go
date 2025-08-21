package system

import (
	"context"
	"fmt"

	"dario.cat/mergo"
	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type SystemModel struct {
	Id        types.StringValue `tfsdk:"id"`
	Type      types.StringValue `tfsdk:"type"`
	Anonymous types.BoolValue   `tfsdk:"anonymous"`

	Hostname        types.StringValue `tfsdk:"hostname"`
	Description     types.StringValue `tfsdk:"description"`
	Notes           types.StringValue `tfsdk:"notes"`
	Buffersize      types.StringValue `tfsdk:"buffersize"`
	ConLogLevel     types.StringValue `tfsdk:"conloglevel"`
	CronLogLevel    types.StringValue `tfsdk:"cronloglevel"`
	KlogconLogLevel types.StringValue `tfsdk:"klogconloglevel"`
	LogBufferSize   types.StringValue `tfsdk:"log_buffer_size"`
	LogFile         types.StringValue `tfsdk:"log_file"`
	LogHostname     types.StringValue `tfsdk:"log_hostname"`
	LogIP           types.StringValue `tfsdk:"log_ip"`
	LogPort         types.StringValue `tfsdk:"log_port"`
	LogPrefix       types.StringValue `tfsdk:"log_prefix"`
	LogProto        types.StringValue `tfsdk:"log_proto"`
	LogRemote       types.StringValue `tfsdk:"log_remote"`
	LogSize         types.StringValue `tfsdk:"log_size"`
	LogTrailerNull  types.StringValue `tfsdk:"log_trailer_null"`
	LogType         types.StringValue `tfsdk:"log_type"`
	TTYLogin        types.StringValue `tfsdk:"ttylogin"`
	UrandomSeed     types.StringValue `tfsdk:"urandom_seed"`
	Timezone        types.StringValue `tfsdk:"timezone"`
	ZoneName        types.StringValue `tfsdk:"zonename"`
	ZramCompAlgo    types.StringValue `tfsdk:"zram_comp_algo"`
	ZramSizeMb      types.StringValue `tfsdk:"zram_size_mb"`
}

// ProjectResource represent Incus project resource.
type SystemResource struct {
	provider api.Client
}

// NewProjectResource return new project resource.
func NewSystemResource() resource.Resource {
	return &SystemResource{}
}

// Metadata for project resource.
func (s SystemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_system", req.ProviderTypeName)
}

// Schema for system resource.
func (s SystemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				CustomType: types.StringType{},
				Computed:   true,
				Optional:   true,
			},

			"anonymous": schema.BoolAttribute{
				CustomType: types.BoolType{},
				Computed:   true,
				Optional:   true,
			},

			"type": schema.StringAttribute{
				CustomType: types.StringType{},
				Computed:   true,
				Optional:   true,
			},

			"hostname": schema.StringAttribute{
				CustomType:  types.StringType{},
				Optional:    true,
				Description: "The hostname for this system",
			},

			"description": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"notes": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"buffersize": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"conloglevel": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"cronloglevel": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"klogconloglevel": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_buffer_size": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_file": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_hostname": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_ip": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_port": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_prefix": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_proto": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_remote": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_size": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_trailer_null": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"log_type": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"ttylogin": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"urandom_seed": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"timezone": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
				// Computed:   true,
				// Default:    stringdefault.StaticString("UTC"),
			},

			"zonename": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
				// Computed:   true,
				// Default:    stringdefault.StaticString("UTC"),
			},

			"zram_comp_algo": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"zram_size_mb": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},
		},
	}
}

func (s *SystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := req.ProviderData
	if data == nil {
		return
	}
	provider, ok := data.(api.Client)
	if !ok {
		resp.Diagnostics.AddError("Failed to get api client", "")
		return
	}
	s.provider = provider
}

func (s SystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name, err := s.provider.Add(ctx, "system", "system")
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create config %q", plan.Id.ValueString()), err.Error())
		return
	}

	plan.Id = types.NewStringValue(name)
	plan.Anonymous = types.NewBoolValue(true)
	plan.Type = types.NewStringValue("system")

	err = s.provider.TSet(ctx, plan, "system", plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", plan.Id.ValueString()), err.Error())
		return
	}

	if err := s.provider.CommitOrRevert(ctx, "system", plan.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to commit or revert", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (s SystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sm, err := s.provider.GetAll(ctx, "system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read config %q", state.Id.ValueString()), err.Error())
		return
	}

	if err := mergo.Merge(&state, sm, mergo.WithOverride, mergo.WithoutDereference); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to merge config system %q", state.Id.ValueString()), err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (s SystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state SystemModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan SystemModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sm, err := s.provider.GetAll(ctx, s.provider, "system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Id.ValueString()), err.Error())
		return
	}

	if err := mergo.Merge(&state, sm, mergo.WithOverride, mergo.WithoutDereference); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Id.ValueString()), err.Error())
		return
	}

	err = s.provider.TSet(ctx, state, "system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Id.ValueString()), err.Error())
		return
	}

	if err := s.provider.CommitOrRevert(ctx, "system", state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to commit or revert", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (s SystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := s.provider.Delete(ctx, "system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete config system config %q", state.Id.ValueString()), err.Error())
		return
	}

	if err := s.provider.CommitOrRevert(ctx, "system", state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to commit or revert", err.Error())
		return
	}
}

func (s *SystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sm, err := s.provider.GetSystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, sm)...)
}
