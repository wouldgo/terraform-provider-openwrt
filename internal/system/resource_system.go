package system

import (
	"context"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// ProjectModel resource data model that matches the schema.
type SystemModel struct {
	Id        types.StringValue `tfsdk:"id" json:".name,omitempty"`
	Type      types.StringValue `tfsdk:"type" json:".type,omitzero,omitempty"`
	Anonymous types.BoolValue   `tfsdk:"anonymous" json:".anonymous,omitzero,omitempty"`

	Hostname        types.StringValue `tfsdk:"hostname" json:"hostname,omitzero"`
	Description     types.StringValue `tfsdk:"description" json:"description,omitzero"`
	Notes           types.StringValue `tfsdk:"notes" json:"notes,omitzero"`
	Buffersize      types.StringValue `tfsdk:"buffersize" json:"buffersize,omitzero"`
	ConLogLevel     types.StringValue `tfsdk:"conloglevel" json:"conloglevel,omitzero"`
	CronLogLevel    types.StringValue `tfsdk:"cronloglevel" json:"cronloglevel,omitzero"`
	KlogconLogLevel types.StringValue `tfsdk:"klogconloglevel" json:"klogconloglevel,omitzero"`
	LogBufferSize   types.StringValue `tfsdk:"log_buffer_size" json:"log_buffer_size,omitzero"`
	LogFile         types.StringValue `tfsdk:"log_file" json:"log_file,omitzero"`
	LogHostname     types.StringValue `tfsdk:"log_hostname" json:"log_hostname,omitzero"`
	LogIP           types.StringValue `tfsdk:"log_ip" json:"log_ip,omitzero"`
	LogPort         types.StringValue `tfsdk:"log_port" json:"log_port,omitzero"`
	LogPrefix       types.StringValue `tfsdk:"log_prefix" json:"log_prefix,omitzero"`
	LogProto        types.StringValue `tfsdk:"log_proto" json:"log_proto,omitzero"`
	LogRemote       types.StringValue `tfsdk:"log_remote" json:"log_remote,omitzero"`
	LogSize         types.StringValue `tfsdk:"log_size" json:"log_size,omitzero"`
	LogTrailerNull  types.StringValue `tfsdk:"log_trailer_null" json:"log_trailer_null,omitzero"`
	LogType         types.StringValue `tfsdk:"log_type" json:"log_type,omitzero"`
	TTYLogin        types.StringValue `tfsdk:"ttylogin" json:"ttylogin,omitzero"`
	UrandomSeed     types.StringValue `tfsdk:"urandom_seed" json:"urandom_seed,omitempty"`
	Timezone        types.StringValue `tfsdk:"timezone" json:"timezone,omitzero"`
	ZoneName        types.StringValue `tfsdk:"zonename" json:"zonename,omitzero"`
	ZramCompAlgo    types.StringValue `tfsdk:"zram_comp_algo" json:"zram_comp_algo,omitzero"`
	ZramSizeMb      types.StringValue `tfsdk:"zram_size_mb" json:"zram_size_mb,omitzero"`
}

// ProjectResource represent Incus project resource.
type SystemResource struct {
	provider *api.Client
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
	provider, ok := data.(*api.Client)
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

	name, err := s.provider.UCIAdd("system", "system")
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create config %q", plan.Id.ValueString()), err.Error())
		return
	}

	plan.Id = types.NewStringValue(name)
	plan.Anonymous = types.NewBoolValue(true)
	plan.Type = types.NewStringValue("system")

	_, err = s.provider.UCITSet(plan, "system", plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", plan.Id.ValueString()), err.Error())
		return
	}

	if _, d := s.provider.UCICommitAndRevert("system", plan.Id.ValueString()); d != nil {
		resp.Diagnostics.Append(d...)
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

	raw, err := s.provider.UCIGetAll("system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read config %q", state.Id.ValueString()), err.Error())
		return
	}
	sm, err := api.Unmarshal[SystemModel](*raw)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read config %q", state.Id.ValueString()), err.Error())
		return
	}

	merged, err := api.Merge[*SystemModel](&state, &sm)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to merge config system %q", state.Id.ValueString()), err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, *merged)...)
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

	sm, err := api.UCIGetAllT[SystemModel](s.provider, "system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Id.ValueString()), err.Error())
		return
	}

	merged, err := api.Merge[SystemModel](&plan, &sm)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Id.ValueString()), err.Error())
		return
	}

	_, err = s.provider.UCITSet(merged, "system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Id.ValueString()), err.Error())
		return
	}

	if _, d := s.provider.UCICommitAndRevert("system", state.Id.ValueString()); d != nil {
		resp.Diagnostics.Append(d...)
		return
	}

	// Transfer the hidden fields
	merged.Anonymous = state.Anonymous
	merged.Id = state.Id
	merged.Type = state.Type

	resp.Diagnostics.Append(resp.State.Set(ctx, merged)...)
}

func (s SystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := s.provider.UCIDelete("system", state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete config system config %q", state.Id.ValueString()), err.Error())
		return
	}

	if _, d := s.provider.UCICommitAndRevert("system", state.Id.ValueString()); d != nil {
		resp.Diagnostics.Append(d...)
		return
	}
}

func (s *SystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw, err := s.provider.UCIGetSystem()
	if err != nil {
		resp.Diagnostics.AddError("Failed to import state", err.Error())
		return
	}
	sm, err := api.Unmarshal[*SystemModel](raw)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sm)...)
}
