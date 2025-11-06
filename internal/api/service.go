// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultListServicesTimeout   time.Duration = 30 * time.Second
	defaultIsEnabledTimeout                    = 30 * time.Second
	defaultDisableServiceTimeout               = 30 * time.Second
	defaultEnableServiceTimeout                = 30 * time.Second
	defaultStartServiceTimeout                 = 30 * time.Second
	defaultStopSeviceTimeout                   = 30 * time.Second
	defaultRestartServiceTimeout               = 30 * time.Second
)

type ServiceTimeouts interface {
	ListServices() time.Duration
	IsEnabled() time.Duration
	DisableService() time.Duration
	EnableService() time.Duration
	StartService() time.Duration
	StopSevice() time.Duration
	RestartService() time.Duration
}

type ServiceTimeoutsModel struct {
	ListServicesTimeout   types.String `tfsdk:"list_services"`
	IsEnabledTimeout      types.String `tfsdk:"is_enabled"`
	DisableServiceTimeout types.String `tfsdk:"disable_service"`
	EnableServiceTimeout  types.String `tfsdk:"enable_service"`
	StartServiceTimeout   types.String `tfsdk:"start_service"`
	StopSeviceTimeout     types.String `tfsdk:"stop_sevice"`
	RestartServiceTimeout types.String `tfsdk:"restart_service"`
}

type ServiceFacade interface {
	ListServices(ctx context.Context) ([]string, error)
	IsEnabled(ctx context.Context, serviceName string) (bool, error)
	DisableService(ctx context.Context, serviceName string) error
	EnableService(ctx context.Context, serviceName string) error
	StartService(ctx context.Context, serviceName string) error
	StopSevice(ctx context.Context, serviceName string) error
	RestartService(ctx context.Context, serviceName string) error
}

type serviceTimeouts struct {
	listServicesTimeout,
	isEnabledTimeout,
	disableServiceTimeout,
	enableServiceTimeout,
	startServiceTimeout,
	stopSeviceTimeout,
	restartServiceTimeout time.Duration
}

func (sT *serviceTimeouts) ListServices() time.Duration {
	return sT.listServicesTimeout
}

func (sT *serviceTimeouts) IsEnabled() time.Duration {
	return sT.isEnabledTimeout
}

func (sT *serviceTimeouts) DisableService() time.Duration {
	return sT.disableServiceTimeout
}

func (sT *serviceTimeouts) EnableService() time.Duration {
	return sT.enableServiceTimeout
}

func (sT *serviceTimeouts) StartService() time.Duration {
	return sT.startServiceTimeout
}

func (sT *serviceTimeouts) StopSevice() time.Duration {
	return sT.stopSeviceTimeout
}

func (sT *serviceTimeouts) RestartService() time.Duration {
	return sT.restartServiceTimeout
}

var (
	_ ServiceFacade   = (*service)(nil)
	_ WithSession     = (*service)(nil)
	_ ServiceTimeouts = (*serviceTimeouts)(nil)

	serviceTimeoutSchemaAttribute = schema.SingleNestedAttribute{
		MarkdownDescription: `Service operations timeout configuration`,
		Description:         `Service operations timeout configuration`,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"list_services": schema.StringAttribute{
				MarkdownDescription: `List services RPC timeout value`,
				Description:         `List services RPC timeout value`,
				Optional:            true,
			},
			"is_enabled": schema.StringAttribute{
				MarkdownDescription: `Is enabled service RPC timeout value`,
				Description:         `Is enabled service RPC timeout value`,
				Optional:            true,
			},
			"disable_service": schema.StringAttribute{
				MarkdownDescription: `Disable service RPC timeout value`,
				Description:         `Disable service RPC timeout value`,
				Optional:            true,
			},
			"enable_service": schema.StringAttribute{
				MarkdownDescription: `Enable service RPC timeout value`,
				Description:         `Enable service RPC timeout value`,
				Optional:            true,
			},
			"start_service": schema.StringAttribute{
				MarkdownDescription: `Start service RPC timeout value`,
				Description:         `Start service RPC timeout value`,
				Optional:            true,
			},
			"stop_sevice": schema.StringAttribute{
				MarkdownDescription: `Stop service RPC timeout value`,
				Description:         `Stop service RPC timeout value`,
				Optional:            true,
			},
			"restart_service": schema.StringAttribute{
				MarkdownDescription: `Restart service RPC timeout value`,
				Description:         `Restart service RPC timeout value`,
				Optional:            true,
			},
		},
	}
)

func parseServiceTimeouts(ctx context.Context, t *TimeoutsModel) (ServiceTimeouts, error) {
	listServicesTimeout := defaultListServicesTimeout
	isEnabledTimeout := defaultIsEnabledTimeout
	disableServiceTimeout := defaultDisableServiceTimeout
	enableServiceTimeout := defaultEnableServiceTimeout
	startServiceTimeout := defaultStartServiceTimeout
	stopSeviceTimeout := defaultStopSeviceTimeout
	restartServiceTimeout := defaultRestartServiceTimeout

	if t != nil && t.Service != nil && !t.Service.ListServicesTimeout.IsNull() {
		parsedListServicesTimeout, err := time.ParseDuration(t.Service.ListServicesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		listServicesTimeout = parsedListServicesTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: list_services config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default list_services config")
	}

	if t != nil && t.Service != nil && !t.Service.IsEnabledTimeout.IsNull() {
		parsedIsEnabledTimeout, err := time.ParseDuration(t.Service.IsEnabledTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		isEnabledTimeout = parsedIsEnabledTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: is_enabled config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default is_enabled config")
	}

	if t != nil && t.Service != nil && !t.Service.DisableServiceTimeout.IsNull() {
		parsedDisableServiceTimeout, err := time.ParseDuration(t.Service.DisableServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		disableServiceTimeout = parsedDisableServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: disable_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default disable_service config")
	}

	if t != nil && t.Service != nil && !t.Service.EnableServiceTimeout.IsNull() {
		parsedEnableServiceTimeout, err := time.ParseDuration(t.Service.EnableServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		enableServiceTimeout = parsedEnableServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: enable_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default enable_service config")
	}

	if t != nil && t.Service != nil && !t.Service.StartServiceTimeout.IsNull() {
		parsedStartServiceTimeout, err := time.ParseDuration(t.Service.StartServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		startServiceTimeout = parsedStartServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: start_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default start_service config")
	}

	if t != nil && t.Service != nil && !t.Service.StopSeviceTimeout.IsNull() {
		parsedStopSeviceTimeout, err := time.ParseDuration(t.Service.StopSeviceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		stopSeviceTimeout = parsedStopSeviceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: stop_sevice config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default stop_sevice config")
	}

	if t != nil && t.Service != nil && !t.Service.RestartServiceTimeout.IsNull() {
		parsedRestartServiceTimeout, err := time.ParseDuration(t.Service.RestartServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		restartServiceTimeout = parsedRestartServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: restart_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default restart_service config")
	}

	return &serviceTimeouts{
		listServicesTimeout,
		isEnabledTimeout,
		disableServiceTimeout,
		enableServiceTimeout,
		startServiceTimeout,
		stopSeviceTimeout,
		restartServiceTimeout,
	}, nil
}

type service struct {
	timeouts ServiceTimeouts

	token  string
	url    *string
	client *http.Client
}

type ServiceInfo struct {
	Version string
	Status  Status
}

func (s *service) SetToken(ctx context.Context, token string) error {
	s.token = token
	return nil
}

func (s *service) ListServices(ctx context.Context) ([]string, error) {
	result, err := call(ctx, s.client, s.timeouts.ListServices(),
		*s.url, s.token,
		"sys", "init.names", []any{})
	if err != nil {
		return nil, err
	}

	var data []string
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, errors.Join(ErrUnMarshal, err)
	}
	return data, nil
}

func (s *service) IsEnabled(ctx context.Context, serviceName string) (bool, error) {
	result, err := call(ctx, s.client, s.timeouts.IsEnabled(),
		*s.url, s.token,
		"sys", "init.enabled", []any{serviceName})
	if err != nil {
		return false, err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return false, errors.Join(ErrUnMarshal, err)
	}
	return data, nil
}

func (s *service) DisableService(ctx context.Context, serviceName string) error {
	result, err := call(ctx, s.client, s.timeouts.DisableService(),
		*s.url, s.token,
		"sys", "init.disable", []any{serviceName})
	if err != nil {
		return err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	if !data {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) EnableService(ctx context.Context, serviceName string) error {
	result, err := call(ctx, s.client, s.timeouts.EnableService(),
		*s.url, s.token,
		"sys", "init.enable", []any{serviceName})
	if err != nil {
		return err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	if !data {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) StartService(ctx context.Context, serviceName string) error {
	result, err := call(ctx, s.client, s.timeouts.StartService(),
		*s.url, s.token,
		"sys", "init.start", []any{serviceName})
	if err != nil {
		return err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}
	if !data {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) StopSevice(ctx context.Context, serviceName string) error {
	result, err := call(ctx, s.client, s.timeouts.StopSevice(),
		*s.url, s.token,
		"sys", "init.stop", []any{serviceName})
	if err != nil {
		return err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}
	if !data {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) RestartService(ctx context.Context, serviceName string) error {
	err := s.StopSevice(ctx, serviceName)
	if err != nil {
		return err
	}
	return s.StartService(ctx, serviceName)
}
