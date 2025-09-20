// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

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

type systemModel struct {
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

type systemResource struct {
	provider api.Client
}

// NewSystemResource return new project resource.
func NewSystemResource() resource.Resource {
	return &systemResource{}
}

// Metadata for project resource.
func (s systemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_system", req.ProviderTypeName)
}

// Schema for system resource.
func (s systemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the system settings in openwrt",
		Description:         "Manage the system settings in openwrt",
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
				MarkdownDescription: "The hostname for this system (Default: \"OpenWrt\")",
				Description:         "The hostname for this system (Default: \"OpenWrt\")",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A short, single-line description for this system. It should be suitable for human consumption in user interfaces, such as LuCI, selector UIs in remote administration applications, or remote UCI (over ubus RPC).",
				Description:         "A short, single-line description for this system. It should be suitable for human consumption in user interfaces, such as LuCI, selector UIs in remote administration applications, or remote UCI (over ubus RPC).",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "A multi-line, free-form text field about this system that can be used in any way the user wishes, e.g. to hold installation notes, or unit serial number and inventory number, location, etc.",
				Description:         "A multi-line, free-form text field about this system that can be used in any way the user wishes, e.g. to hold installation notes, or unit serial number and inventory number, location, etc.",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"buffersize": schema.StringAttribute{
				MarkdownDescription: "Size of the kernel message buffer.",
				Description:         "Size of the kernel message buffer.",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"conloglevel": schema.StringAttribute{
				MarkdownDescription: "The maximum log level for kernel messages to be logged to the console. (Default: 7)",
				Description:         "The maximum log level for kernel messages to be logged to the console. (Default: 7)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"cronloglevel": schema.StringAttribute{
				MarkdownDescription: "The minimum level for cron messages to be logged to syslog. 0 will print all debug messages, 8 will log command executions, and 9 or higher will only log error messages. (Default: 5)",
				Description:         "The minimum level for cron messages to be logged to syslog. 0 will print all debug messages, 8 will log command executions, and 9 or higher will only log error messages. (Default: 5)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"klogconloglevel": schema.StringAttribute{
				MarkdownDescription: "The maximum log level for kernel messages to be logged to the console. Only messages with a level lower than this will be printed to the console. Identical to conloglevel and will override it. (Default: 7)",
				Description:         "The maximum log level for kernel messages to be logged to the console. Only messages with a level lower than this will be printed to the console. Identical to conloglevel and will override it. (Default: 7)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_buffer_size": schema.StringAttribute{
				MarkdownDescription: "Size of the log buffer of the procd based system log, that is accessible via the logread command. Defaults to the value of log_size if unset.",
				Description:         "Size of the log buffer of the procd based system log, that is accessible via the logread command. Defaults to the value of log_size if unset.",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_file": schema.StringAttribute{
				MarkdownDescription: "File to write log messages to (type file). The default is to not write a log in a file. The most often used location for a system log file is `/var/log/messages`.",
				Description:         "File to write log messages to (type file). The default is to not write a log in a file. The most often used location for a system log file is /var/log/messages.",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname to send to remote syslog. If none is provided, the actual hostname is send. This feature is only present in 17.xx and later versions",
				Description:         "Hostname to send to remote syslog. If none is provided, the actual hostname is send. This feature is only present in 17.xx and later versions",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_ip": schema.StringAttribute{
				MarkdownDescription: "IP address of a syslog server to which the log messages should be sent in addition to the local destination.",
				Description:         "IP address of a syslog server to which the log messages should be sent in addition to the local destination.",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_port": schema.StringAttribute{
				MarkdownDescription: "Port number of the remote syslog server specified with log_ip. (Default: 514)",
				Description:         "Port number of the remote syslog server specified with log_ip. (Default: 514)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_prefix": schema.StringAttribute{
				MarkdownDescription: "Adds a prefix to all log messages send over network.",
				Description:         "Adds a prefix to all log messages send over network.",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_proto": schema.StringAttribute{
				MarkdownDescription: "Sets the protocol to use for the connection, either tcp or udp. (Default: \"udp\")",
				Description:         "Sets the protocol to use for the connection, either tcp or udp. (Default: \"udp\")",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_remote": schema.StringAttribute{
				MarkdownDescription: "Enables remote logging. (Default: 1)",
				Description:         "Enables remote logging. (Default: 1)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_size": schema.StringAttribute{
				MarkdownDescription: "Size of the file based log buffer in KiB (see log_file). This value is used as the fallback value for log_buffer_size if the latter is not specified. (Default: 64)",
				Description:         "Size of the file based log buffer in KiB (see log_file). This value is used as the fallback value for log_buffer_size if the latter is not specified. (Default: 64)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_trailer_null": schema.StringAttribute{
				MarkdownDescription: "Use \\0 instead of \\n as trailer when using TCP. (Default: 0)",
				Description:         "Use \\0 instead of \\n as trailer when using TCP. (Default: 0)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"log_type": schema.StringAttribute{
				MarkdownDescription: "Either circular or file. The circular option is a fixed size queue in memory, while the file is a dynamically sized file, that can be in memory, or written to disk. Note: If log_type is set to file, then at some point when the log fills, the device may encounter an out-of-space condition. This is especially an issue for devices with limited onboard storage: in memory, or on flash. (Default: \"circular\")",
				Description:         "Either circular or file. The circular option is a fixed size queue in memory, while the file is a dynamically sized file, that can be in memory, or written to disk. Note: If log_type is set to file, then at some point when the log fills, the device may encounter an out-of-space condition. This is especially an issue for devices with limited onboard storage: in memory, or on flash. (Default: \"circular\")",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"ttylogin": schema.StringAttribute{
				MarkdownDescription: "Require authentication for local users to log in the system. Disabled by default. It applies to the access methods listed in /etc/inittab, such as keyboard and serial. (Default: 0)",
				Description:         "Require authentication for local users to log in the system. Disabled by default. It applies to the access methods listed in /etc/inittab, such as keyboard and serial. (Default: 0)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"urandom_seed": schema.StringAttribute{
				MarkdownDescription: "Path of the seed. Enables saving a new seed on each boot. (Default: 0)",
				Description:         "Path of the seed. Enables saving a new seed on each boot. (Default: 0)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "POSIX.1 time zone string corresponding to the time zone in which date and time should be displayed by default. See [timezone database](https://github.com/openwrt/luci/blob/master/modules/luci-lua-runtime/luasrc/sys/zoneinfo/tzdata.lua) for a mapping between IANA/Olson and POSIX.1 formats. (For London this corresponds to GMT0BST,M3.5.0/1,M10.5.0) (Default: \"UTC\")",
				Description:         "POSIX.1 time zone string corresponding to the time zone in which date and time should be displayed by default. See [timezone database](https://github.com/openwrt/luci/blob/master/modules/luci-lua-runtime/luasrc/sys/zoneinfo/tzdata.lua) for a mapping between IANA/Olson and POSIX.1 formats. (For London this corresponds to GMT0BST,M3.5.0/1,M10.5.0) (Default: \"UTC\")",
				CustomType:          types.StringType{},
				Optional:            true,
				// Computed:   true,
				// Default:    stringdefault.StaticString("UTC"),
			},
			"zonename": schema.StringAttribute{
				MarkdownDescription: "IANA/Olson time zone string. If zoneinfo-* packages are present, possible values can be found by running find /usr/share/zoneinfo. See [timezone database](https://github.com/openwrt/luci/blob/master/modules/luci-lua-runtime/luasrc/sys/zoneinfo/tzdata.lua) for a mapping between IANA/Olson and POSIX.1 formats. (For London this corresponds to Europe/London) (Default: UTC)",
				Description:         "IANA/Olson time zone string. If zoneinfo-* packages are present, possible values can be found by running find /usr/share/zoneinfo. See [timezone database](https://github.com/openwrt/luci/blob/master/modules/luci-lua-runtime/luasrc/sys/zoneinfo/tzdata.lua) for a mapping between IANA/Olson and POSIX.1 formats. (For London this corresponds to Europe/London) (Default: UTC)",
				CustomType:          types.StringType{},
				Optional:            true,
				// Computed:   true,
				// Default:    stringdefault.StaticString("UTC"),
			},
			"zram_comp_algo": schema.StringAttribute{
				MarkdownDescription: "Compression algorithm to use for ZRAM, can be one of lzo, lzo-rle, lz4, zstd. (Default: \"lzo\")",
				Description:         "Compression algorithm to use for ZRAM, can be one of lzo, lzo-rle, lz4, zstd. (Default: \"lzo\")",
				CustomType:          types.StringType{},
				Optional:            true,
			},
			"zram_size_mb": schema.StringAttribute{
				MarkdownDescription: "Size of ZRAM in MB. (Default: ramsize in Kb divided by 2048)",
				Description:         "Size of ZRAM in MB. (Default: ramsize in Kb divided by 2048)",
				CustomType:          types.StringType{},
				Optional:            true,
			},
		},
	}
}

func (s *systemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (s systemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemModel
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

func (s systemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemModel
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

func (s systemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state systemModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan systemModel
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

func (s systemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state systemModel
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

func (s *systemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sm, err := s.provider.GetSystem(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, sm)...)
}
