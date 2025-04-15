package system

import (
	"context"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/foxboron/terraform-provider-openwrt/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
)

// ProjectModel resource data model that matches the schema.
type SystemModel struct {
	Name            types.StringValue `tfsdk:"name" json:".name,omitempty"`
	Anonymous       types.BoolValue   `tfsdk:"anonymous" json:".anonymous,omitzero,omitempty"`
	Type            types.StringValue `tfsdk:"type" json:".type,omitzero,omitempty"`
	ZoneName        types.StringValue `tfsdk:"zonename" json:"zonename,omitzero"`
	Timezone        types.StringValue `tfsdk:"timezone" json:"timezone,omitzero"`
	LogSize         types.StringValue `tfsdk:"log_size" json:"log_size,omitzero"`
	Hostname        types.StringValue `tfsdk:"hostname" json:"hostname,omitzero"`
	TTYLogin        types.StringValue `tfsdk:"ttylogin" json:"ttylogin,omitzero"`
	ConLogLevel     types.StringValue `tfsdk:"conloglevel" json:"conloglevel,omitzero"`
	CronLogLevel    types.StringValue `tfsdk:"cronloglevel" json:"cronloglevel,omitzero"`
	KlogconLogLevel types.StringValue `tfsdk:"klogconloglevel" json:"klogconloglevel,omitzero"`
	UrandomSeed     types.StringValue `tfsdk:"urandom_seed" json:"urandom_seed,omitempty"`
	Description     types.StringValue `tfsdk:"description" json:"description,omitzero"`
	Notes           types.StringValue `tfsdk:"notes" json:"notes,omitzero"`
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
			"anonymous": schema.BoolAttribute{
				CustomType: types.BoolType{},
				Computed:   true,
				Optional:   true,
			},

			"name": schema.StringAttribute{
				CustomType: types.StringType{},
				Computed:   true,
				Optional:   true,
			},

			"type": schema.StringAttribute{
				CustomType: types.StringType{},
				Computed:   true,
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

			"ttylogin": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"urandom_seed": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			// "log_buffer_size": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_file": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_hostname": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_ip": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_port": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_prefix": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_proto": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },
			//
			// "log_remote": schema.StringAttribute{
			// 	CustomType: types.StringType{},
			// 	Optional:   true,
			// },

			"log_size": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"timezone": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString("UTC"),
			},

			"zonename": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString("UTC"),
			},

			"hostname": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"description": schema.StringAttribute{
				CustomType: types.StringType{},
				Optional:   true,
			},

			"notes": schema.StringAttribute{
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
		fmt.Println("failed getting api client")
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
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create config %q", plan.Name.ValueString()), err.Error())
		return
	}

	plan.Name = types.NewStringValue(name)
	plan.Anonymous = types.NewBoolValue(true)
	plan.Type = types.NewStringValue("system")

	_, err = s.provider.UCITSet(plan, "system", plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", plan.Name.ValueString()), err.Error())
		return
	}

	if _, d := s.provider.UCICommitAndRevert("system", plan.Name.ValueString()); d != nil {
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

	raw, err := s.provider.UCIGetAll("system", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read config %q", state.Name.ValueString()), err.Error())
		return
	}
	sm, err := api.Unmarshal[SystemModel](*raw)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to read config %q", state.Name.ValueString()), err.Error())
		return
	}

	merged, err := api.Merge[*SystemModel](&sm, &state)
	if err != nil {
		panic(err)
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

	sm, err := api.UCIGetAllT[SystemModel](s.provider, "system", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Name.ValueString()), err.Error())
		return
	}

	merged, err := api.Merge[SystemModel](&plan, &sm)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Name.ValueString()), err.Error())
		return
	}

	_, err = s.provider.UCITSet(merged, "system", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update config %q", state.Name.ValueString()), err.Error())
		return
	}

	if _, d := s.provider.UCICommitAndRevert("system", state.Name.ValueString()); d != nil {
		resp.Diagnostics.Append(d...)
		return
	}

	// Transfer the hidden fields
	merged.Anonymous = state.Anonymous
	merged.Name = state.Name
	merged.Type = state.Type

	resp.Diagnostics.Append(resp.State.Set(ctx, merged)...)
}

func (s SystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fmt.Println("calling delete")
	var state SystemModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := s.provider.UCIDelete("system", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete config system config %q", state.Name.ValueString()), err.Error())
		return
	}

	if _, d := s.provider.UCICommitAndRevert("system", state.Name.ValueString()); d != nil {
		resp.Diagnostics.Append(d...)
		return
	}
}

func (s *SystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw, err := s.provider.UCIGetSystem()
	if err != nil {
		panic(err)
	}
	sm, err := api.Unmarshal[*SystemModel](raw)
	if err != nil {
		panic(err)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, sm)...)
}
