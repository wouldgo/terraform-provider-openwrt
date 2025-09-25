// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceModel struct {
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Triggers types.Map    `tfsdk:"triggers"`
}

type serviceResource struct {
	initFacade api.InitFacade
}

func NewServiceResource() resource.Resource {
	return &serviceResource{}
}

func (s serviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_service", req.ProviderTypeName)
}

func (s serviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Enable/Disable specific services on the router, verifing conditions that trigger the restart on the router",
		Description:         "Enable/Disable specific services on the router, verifing conditions that trigger the restart on the router",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The service name to operate with",
				Description:         "The service name to operate with",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the service must be enabled",
				Description:         "Whether the service must be enabled",
				Optional:            true,
				Computed:            true,
			},
			"triggers": schema.MapAttribute{
				MarkdownDescription: "Key/value map that forces update when changed",
				Description:         "Key/value map that forces update when changed",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (s *serviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data := req.ProviderData
	if data == nil {
		return
	}
	initFacade, ok := data.(api.InitFacade)
	if !ok {
		resp.Diagnostics.AddError("failed to get init facace", "")
		return
	}
	s.initFacade = initFacade
}

func (s serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Enabled.IsUnknown() || plan.Enabled.IsNull() {
		plan.Enabled = types.BoolValue(true)
	}

	if err := s.enableDisableService(ctx,
		plan.Enabled, types.BoolNull(),
		plan.Triggers, types.MapNull(types.StringType),
		plan.Name.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("failed to create resource", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (s serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := state.Name.ValueString()
	enabled, err := s.initFacade.IsEnabled(ctx, serviceName)
	if err != nil {
		resp.Diagnostics.AddError("checking if service is enabled in error", fmt.Sprintf("%s: %v", serviceName, err))
		return
	}
	state.Enabled = types.BoolValue(enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (s serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state serviceModel
	var plan serviceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := s.enableDisableService(ctx,
		plan.Enabled, state.Enabled,
		plan.Triggers, state.Triggers,
		plan.Name.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("failed to update resource", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (s serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := state.Name.ValueString()
	toEnable := state.Enabled.ValueBool()

	if toEnable {
		if err := s.initFacade.EnableService(ctx, serviceName); err != nil {
			resp.Diagnostics.AddError("failed to enable service", fmt.Sprintf("%s: %v", serviceName, err))
			return
		}
	} else {
		if err := s.initFacade.DisableService(ctx, serviceName); err != nil {
			resp.Diagnostics.AddError("failed to disable service", fmt.Sprintf("%s: %v", serviceName, err))
			return
		}
	}
}

func (s serviceResource) enableDisableService(ctx context.Context,
	planEnabledValue, stateEnabledValue types.Bool,
	planTriggersValue, stateTriggersValue types.Map,
	serviceName string) error {
	toEnable := planEnabledValue.ValueBool()
	if !planEnabledValue.Equal(stateEnabledValue) {
		if toEnable {
			if err := s.initFacade.EnableService(ctx, serviceName); err != nil {
				return fmt.Errorf("failed to enable service: %w", err)
			}
		} else {
			if err := s.initFacade.DisableService(ctx, serviceName); err != nil {
				return fmt.Errorf("failed to disable service: %w", err)
			}
		}
	}

	if toEnable && !planTriggersValue.Equal(stateTriggersValue) {
		if err := s.initFacade.RestartService(ctx, serviceName); err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}
	}
	return nil
}
