// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/fs"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/opkg"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/service"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/system"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const defaultAuthTimeout = 5 * time.Second

type TimeoutsModel struct {
	Auth types.String `tfsdk:"auth"`

	Fs      *fs.FsTimeoutsModel           `tfsdk:"fs"`
	Opkg    *opkg.OpkgTimeoutsModel       `tfsdk:"opkg"`
	Service *service.ServiceTimeoutsModel `tfsdk:"service"`
	System  *system.SystemTimeoutsModel   `tfsdk:"uci"`
}

var (
	_ rpc.Timeouts = (*timeouts)(nil)
)

type timeouts struct {
	rpc.FsTimeouts
	rpc.OpkgTimeouts
	rpc.ServiceTimeouts
	rpc.SystemTimeouts

	authTimeout time.Duration
}

func (t *timeouts) Auth() time.Duration {
	return t.authTimeout
}

func parseTimeouts(ctx context.Context, t *TimeoutsModel) (rpc.Timeouts, error) {
	authTimeout := defaultAuthTimeout

	var (
		fsTimeouts      rpc.FsTimeouts
		opkgTimeouts    rpc.OpkgTimeouts
		serviceTimeouts rpc.ServiceTimeouts
		systemTimeouts  rpc.SystemTimeouts
		err             error
	)

	if t != nil && !t.Auth.IsNull() {
		parsedAuthTimeout, err := time.ParseDuration(t.Auth.ValueString())
		if err != nil {
			return nil, err
		}

		authTimeout = parsedAuthTimeout
		tflog.Debug(ctx, "parse timeout configuration: auth config parsed")
	} else {
		tflog.Debug(ctx, "parse timeout configuration: default auth config")
	}

	if t != nil && t.Fs != nil {
		fsTimeouts, err = fs.ParseTimeouts(ctx, t.Fs)
		if err != nil {
			return nil, fmt.Errorf("error parsing fs timeouts: %w", err)
		}
	} else {
		fsTimeouts = fs.DefaultTimeouts
	}

	if t != nil && t.Opkg != nil {
		opkgTimeouts, err = opkg.ParseTimeouts(ctx, t.Opkg)
		if err != nil {
			return nil, fmt.Errorf("error parsing opkg timeouts: %w", err)
		}
	} else {
		opkgTimeouts = opkg.DefaultTimeouts
	}

	if t != nil && t.Service != nil {
		serviceTimeouts, err = service.ParseTimeouts(ctx, t.Service)
		if err != nil {
			return nil, fmt.Errorf("error parsing service timeouts: %w", err)
		}
	} else {
		serviceTimeouts = service.DefaultTimeouts
	}

	if t != nil && t.System != nil {
		systemTimeouts, err = system.ParseTimeouts(ctx, t.System)
		if err != nil {
			return nil, fmt.Errorf("error parsing system timeouts: %w", err)
		}
	} else {
		systemTimeouts = system.DefaultTimeouts
	}

	return &timeouts{
		fsTimeouts,
		opkgTimeouts,
		serviceTimeouts,
		systemTimeouts,
		authTimeout,
	}, nil
}
