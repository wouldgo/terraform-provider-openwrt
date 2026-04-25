// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/fs"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/opkg"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/service"
	"github.com/foxboron/terraform-provider-openwrt/internal/resources/system"
)

type AttemptsModel struct {
	Fs      *fs.FsAttemptsModel           `tfsdk:"fs"`
	Opkg    *opkg.OpkgAttemptsModel       `tfsdk:"opkg"`
	Service *service.ServiceAttemptsModel `tfsdk:"service"`
	System  *system.SystemAttemptsModel   `tfsdk:"uci"`
}

var (
	_ rpc.Attempts = (*attempts)(nil)
)

type attempts struct {
	rpc.FsAttempts
	rpc.OpkgAttempts
	rpc.ServiceAttempts
	rpc.SystemAttempts
}

func parseAttempts(ctx context.Context, r *AttemptsModel) (rpc.Attempts, error) {
	var (
		fsAttempts      rpc.FsAttempts
		opkgAttempts    rpc.OpkgAttempts
		serviceAttempts rpc.ServiceAttempts
		systemAttempts  rpc.SystemAttempts
		err             error
	)

	if r != nil && r.Fs != nil {
		fsAttempts, err = fs.ParseAttempts(ctx, r.Fs)
		if err != nil {
			return nil, fmt.Errorf("error parsing fs attempts: %w", err)
		}
	} else {
		fsAttempts = fs.DefaultAttempts
	}

	if r != nil && r.Opkg != nil {
		opkgAttempts, err = opkg.ParseAttempts(ctx, r.Opkg)
		if err != nil {
			return nil, fmt.Errorf("error parsing opkg attempts: %w", err)
		}
	} else {
		opkgAttempts = opkg.DefaultAttempts
	}

	if r != nil && r.Service != nil {
		serviceAttempts, err = service.ParseAttempts(ctx, r.Service)
		if err != nil {
			return nil, fmt.Errorf("error parsing service attempts: %w", err)
		}
	} else {
		serviceAttempts = service.DefaultAttempts
	}

	if r != nil && r.System != nil {
		systemAttempts, err = system.ParseAttempts(ctx, r.System)
		if err != nil {
			return nil, fmt.Errorf("error parsing system attempts: %w", err)
		}
	} else {
		systemAttempts = system.DefaultAttempts
	}

	return &attempts{
		fsAttempts,
		opkgAttempts,
		serviceAttempts,
		systemAttempts,
	}, nil
}
