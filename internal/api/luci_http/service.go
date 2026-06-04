// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http

import (
	"context"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/http"
	http_transformers "github.com/foxboron/terraform-provider-openwrt/internal/http/transformers"
)

const (
	serviceRPC           = "sys"
	serviceMethodNames   = "init.names"
	serviceMethodEnabled = "init.enabled"
	serviceMethodDisable = "init.disable"
	serviceMethodEnable  = "init.enable"
	serviceMethodStart   = "init.start"
	serviceMethodStop    = "init.stop"
)

var (
	_ luci.ServiceFacade = (*service)(nil)
)

type service struct {
	*http.BaseClient
	timeouts luci.ServiceTimeouts
}

func (s *service) ListServices(ctx context.Context) ([]string, error) {
	return http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.ListServices(),
		serviceRPC,
		serviceMethodNames,
		http_transformers.StringSliceTransformer,
	)
}

func (s *service) IsEnabled(ctx context.Context, serviceName string) (bool, error) {
	return http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.IsEnabled(),
		serviceRPC,
		serviceMethodEnabled,
		http_transformers.BooleanTransformer,
		serviceName,
	)
}

func (s *service) DisableService(ctx context.Context, serviceName string) error {
	result, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.DisableService(),
		serviceRPC,
		serviceMethodDisable,
		http_transformers.BooleanTransformer,
		serviceName,
	)

	if err != nil {
		return fmt.Errorf("unable to disable service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to disable service %s: %w", serviceName, luci.ErrExecutionFailure)
	}
	return nil
}

func (s *service) EnableService(ctx context.Context, serviceName string) error {
	result, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.EnableService(),
		serviceRPC,
		serviceMethodEnable,
		http_transformers.BooleanTransformer,
		serviceName,
	)
	if err != nil {
		return fmt.Errorf("unable to enable service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to enable service %s: %w", serviceName, luci.ErrExecutionFailure)
	}
	return nil
}

func (s *service) StartService(ctx context.Context, serviceName string) error {
	result, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.StartService(),
		serviceRPC,
		serviceMethodStart,
		http_transformers.BooleanTransformer,
		serviceName,
	)
	if err != nil {
		return fmt.Errorf("unable to start service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to start service %s: %w", serviceName, luci.ErrExecutionFailure)
	}
	return nil
}

func (s *service) StopSevice(ctx context.Context, serviceName string) error {
	result, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.StopSevice(),
		serviceRPC,
		serviceMethodStop,
		http_transformers.BooleanTransformer,
		serviceName,
	)
	if err != nil {
		return fmt.Errorf("unable to stop service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to stop service %s: %w", serviceName, luci.ErrExecutionFailure)
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
