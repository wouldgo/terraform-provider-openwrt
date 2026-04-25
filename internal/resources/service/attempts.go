// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultListServicesAttempts   int32 = 1
	defaultIsEnabledAttempts      int32 = 1
	defaultDisableServiceAttempts int32 = 1
	defaultEnableServiceAttempts  int32 = 1
	defaultStartServiceAttempts   int32 = 1
	defaultStopSeviceAttempts     int32 = 1
	defaultRestartServiceAttempts int32 = 1
)

var (
	_               rpc.ServiceAttempts = serviceAttempts{}
	DefaultAttempts                     = serviceAttempts{
		defaultListServicesAttempts,
		defaultIsEnabledAttempts,
		defaultDisableServiceAttempts,
		defaultEnableServiceAttempts,
		defaultStartServiceAttempts,
		defaultStopSeviceAttempts,
		defaultRestartServiceAttempts,
	}
)

type serviceAttempts struct {
	listServicesAttempts,
	isEnabledAttempts,
	disableServiceAttempts,
	enableServiceAttempts,
	startServiceAttempts,
	stopSeviceAttempts,
	restartServiceAttempts int32
}

func (sR serviceAttempts) ListServices() int32 {
	return sR.listServicesAttempts
}

func (sR serviceAttempts) IsEnabled() int32 {
	return sR.isEnabledAttempts
}

func (sR serviceAttempts) DisableService() int32 {
	return sR.disableServiceAttempts
}

func (sR serviceAttempts) EnableService() int32 {
	return sR.enableServiceAttempts
}

func (sR serviceAttempts) StartService() int32 {
	return sR.startServiceAttempts
}

func (sR serviceAttempts) StopSevice() int32 {
	return sR.stopSeviceAttempts
}

func (sR serviceAttempts) RestartService() int32 {
	return sR.restartServiceAttempts
}

func ParseAttempts(ctx context.Context, t *ServiceAttemptsModel) (rpc.ServiceAttempts, error) {
	listServicesAttempts := defaultListServicesAttempts
	isEnabledAttempts := defaultIsEnabledAttempts
	disableServiceAttempts := defaultDisableServiceAttempts
	enableServiceAttempts := defaultEnableServiceAttempts
	startServiceAttempts := defaultStartServiceAttempts
	stopSeviceAttempts := defaultStopSeviceAttempts
	restartServiceAttempts := defaultRestartServiceAttempts

	if t != nil && !t.ListServicesAttempts.IsNull() {
		parsedListServicesAttempts := t.ListServicesAttempts.ValueInt32()

		listServicesAttempts = parsedListServicesAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: list_services config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default list_services config")
	}

	if t != nil && !t.IsEnabledAttempts.IsNull() {
		parsedIsEnabledAttempts := t.IsEnabledAttempts.ValueInt32()

		isEnabledAttempts = parsedIsEnabledAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: is_enabled config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default is_enabled config")
	}

	if t != nil && !t.DisableServiceAttempts.IsNull() {
		parsedDisableServiceAttempts := t.DisableServiceAttempts.ValueInt32()

		disableServiceAttempts = parsedDisableServiceAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: disable_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default disable_service config")
	}

	if t != nil && !t.EnableServiceAttempts.IsNull() {
		parsedEnableServiceAttempts := t.EnableServiceAttempts.ValueInt32()

		enableServiceAttempts = parsedEnableServiceAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: enable_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default enable_service config")
	}

	if t != nil && !t.StartServiceAttempts.IsNull() {
		parsedStartServiceAttempts := t.StartServiceAttempts.ValueInt32()

		startServiceAttempts = parsedStartServiceAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: start_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default start_service config")
	}

	if t != nil && !t.StopSeviceAttempts.IsNull() {
		parsedStopSeviceAttempts := t.StopSeviceAttempts.ValueInt32()

		stopSeviceAttempts = parsedStopSeviceAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: stop_sevice config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default stop_sevice config")
	}

	if t != nil && !t.RestartServiceAttempts.IsNull() {
		parsedRestartServiceAttempts := t.RestartServiceAttempts.ValueInt32()

		restartServiceAttempts = parsedRestartServiceAttempts
		tflog.Debug(ctx, "service - parse attempts configuration: restart_service config parsed")
	} else {
		tflog.Debug(ctx, "service - parse attempts configuration: default restart_service config")
	}

	toReturn := serviceAttempts{
		listServicesAttempts,
		isEnabledAttempts,
		disableServiceAttempts,
		enableServiceAttempts,
		startServiceAttempts,
		stopSeviceAttempts,
		restartServiceAttempts,
	}
	tflog.Debug(ctx, "service - attempts configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
