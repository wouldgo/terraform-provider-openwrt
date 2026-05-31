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
	rawResult, err := s.Call(
		ctx,
		s.timeouts.ListServices(),
		serviceRPC,
		serviceMethodNames,
	)
	if err != nil {
		return nil, err
	}

	return http_transformers.StringSliceTransformer.Transform(rawResult)
}

func (s *service) IsEnabled(ctx context.Context, serviceName string) (bool, error) {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.IsEnabled(),
		serviceRPC,
		serviceMethodEnabled,
		serviceName,
	)
	if err != nil {
		return false, err
	}

	return http_transformers.BooleanTransformer.Transform(rawResult)
}

func (s *service) DisableService(ctx context.Context, serviceName string) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.DisableService(),
		serviceRPC,
		serviceMethodDisable,
		serviceName,
	)
	if err != nil {
		return err
	}

	result, err := http_transformers.BooleanTransformer.Transform(rawResult)
	if err != nil {
		return fmt.Errorf("unable to disable service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to disable service %s: %w", serviceName, luci.ErrExecutionFailure)
	}
	return nil
}

func (s *service) EnableService(ctx context.Context, serviceName string) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.EnableService(),
		serviceRPC,
		serviceMethodEnable,
		serviceName,
	)
	if err != nil {
		return err
	}

	result, err := http_transformers.BooleanTransformer.Transform(rawResult)
	if err != nil {
		return fmt.Errorf("unable to enable service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to enable service %s: %w", serviceName, luci.ErrExecutionFailure)
	}
	return nil
}

func (s *service) StartService(ctx context.Context, serviceName string) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.StartService(),
		serviceRPC,
		serviceMethodStart,
		serviceName,
	)
	if err != nil {
		return err
	}

	result, err := http_transformers.BooleanTransformer.Transform(rawResult)
	if err != nil {
		return fmt.Errorf("unable to start service %s: %w", serviceName, err)
	}

	if !result {
		return fmt.Errorf("unable to start service %s: %w", serviceName, luci.ErrExecutionFailure)
	}
	return nil
}

func (s *service) StopSevice(ctx context.Context, serviceName string) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.StopSevice(),
		serviceRPC,
		serviceMethodStop,
		serviceName,
	)
	if err != nil {
		return err
	}

	result, err := http_transformers.BooleanTransformer.Transform(rawResult)
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
