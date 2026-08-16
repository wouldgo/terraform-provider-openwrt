// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"
	"time"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultListServicesTimeout   = 30 * time.Second
	defaultIsEnabledTimeout      = 30 * time.Second
	defaultDisableServiceTimeout = 30 * time.Second
	defaultEnableServiceTimeout  = 30 * time.Second
	defaultStartServiceTimeout   = 30 * time.Second
	defaultStopSeviceTimeout     = 30 * time.Second
	defaultRestartServiceTimeout = 30 * time.Second
)

var (
	_               rpc.ServiceTimeouts = serviceTimeouts{}
	DefaultTimeouts                     = serviceTimeouts{
		defaultListServicesTimeout,
		defaultIsEnabledTimeout,
		defaultDisableServiceTimeout,
		defaultEnableServiceTimeout,
		defaultStartServiceTimeout,
		defaultStopSeviceTimeout,
		defaultRestartServiceTimeout,
	}
)

type serviceTimeouts struct {
	listServicesTimeout,
	isEnabledTimeout,
	disableServiceTimeout,
	enableServiceTimeout,
	startServiceTimeout,
	stopSeviceTimeout,
	restartServiceTimeout time.Duration
}

func (sT serviceTimeouts) ListServices() time.Duration {
	return sT.listServicesTimeout
}

func (sT serviceTimeouts) IsEnabled() time.Duration {
	return sT.isEnabledTimeout
}

func (sT serviceTimeouts) DisableService() time.Duration {
	return sT.disableServiceTimeout
}

func (sT serviceTimeouts) EnableService() time.Duration {
	return sT.enableServiceTimeout
}

func (sT serviceTimeouts) StartService() time.Duration {
	return sT.startServiceTimeout
}

func (sT serviceTimeouts) StopService() time.Duration {
	return sT.stopSeviceTimeout
}

func (sT serviceTimeouts) RestartService() time.Duration {
	return sT.restartServiceTimeout
}

func ParseTimeouts(ctx context.Context, t *ServiceTimeoutsModel) (rpc.ServiceTimeouts, error) {
	listServicesTimeout := defaultListServicesTimeout
	isEnabledTimeout := defaultIsEnabledTimeout
	disableServiceTimeout := defaultDisableServiceTimeout
	enableServiceTimeout := defaultEnableServiceTimeout
	startServiceTimeout := defaultStartServiceTimeout
	stopSeviceTimeout := defaultStopSeviceTimeout
	restartServiceTimeout := defaultRestartServiceTimeout

	if t != nil && !t.ListServicesTimeout.IsNull() {
		parsedListServicesTimeout, err := time.ParseDuration(t.ListServicesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		listServicesTimeout = parsedListServicesTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: list_services config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default list_services config")
	}

	if t != nil && !t.IsEnabledTimeout.IsNull() {
		parsedIsEnabledTimeout, err := time.ParseDuration(t.IsEnabledTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		isEnabledTimeout = parsedIsEnabledTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: is_enabled config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default is_enabled config")
	}

	if t != nil && !t.DisableServiceTimeout.IsNull() {
		parsedDisableServiceTimeout, err := time.ParseDuration(t.DisableServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		disableServiceTimeout = parsedDisableServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: disable_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default disable_service config")
	}

	if t != nil && !t.EnableServiceTimeout.IsNull() {
		parsedEnableServiceTimeout, err := time.ParseDuration(t.EnableServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		enableServiceTimeout = parsedEnableServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: enable_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default enable_service config")
	}

	if t != nil && !t.StartServiceTimeout.IsNull() {
		parsedStartServiceTimeout, err := time.ParseDuration(t.StartServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		startServiceTimeout = parsedStartServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: start_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default start_service config")
	}

	if t != nil && !t.StopSeviceTimeout.IsNull() {
		parsedStopSeviceTimeout, err := time.ParseDuration(t.StopSeviceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		stopSeviceTimeout = parsedStopSeviceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: stop_sevice config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default stop_sevice config")
	}

	if t != nil && !t.RestartServiceTimeout.IsNull() {
		parsedRestartServiceTimeout, err := time.ParseDuration(t.RestartServiceTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		restartServiceTimeout = parsedRestartServiceTimeout
		tflog.Debug(ctx, "service - parse timeout configuration: restart_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse timeout configuration: default restart_service config")
	}

	toReturn := serviceTimeouts{
		listServicesTimeout,
		isEnabledTimeout,
		disableServiceTimeout,
		enableServiceTimeout,
		startServiceTimeout,
		stopSeviceTimeout,
		restartServiceTimeout,
	}
	tflog.Debug(ctx, "service - timeout configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
