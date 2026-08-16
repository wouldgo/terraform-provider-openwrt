// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"
)

type ServiceFacade interface {
	ListServices(ctx context.Context) ([]string, error)
	IsEnabled(ctx context.Context, serviceName string) (bool, error)
	DisableService(ctx context.Context, serviceName string) error
	EnableService(ctx context.Context, serviceName string) error
	StartService(ctx context.Context, serviceName string) error
	StopService(ctx context.Context, serviceName string) error
	RestartService(ctx context.Context, serviceName string) error
}

type ServiceTimeouts interface {
	ListServices() time.Duration
	IsEnabled() time.Duration
	DisableService() time.Duration
	EnableService() time.Duration
	StartService() time.Duration
	StopService() time.Duration
	RestartService() time.Duration
}
