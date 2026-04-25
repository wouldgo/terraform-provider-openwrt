// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
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

type ServiceFacade interface {
	ListServices(ctx context.Context) ([]string, error)
	IsEnabled(ctx context.Context, serviceName string) (bool, error)
	DisableService(ctx context.Context, serviceName string) error
	EnableService(ctx context.Context, serviceName string) error
	StartService(ctx context.Context, serviceName string) error
	StopSevice(ctx context.Context, serviceName string) error
	RestartService(ctx context.Context, serviceName string) error
}

var (
	_ ServiceFacade = (*service)(nil)
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

type ServiceAttempts interface {
	ListServices() int32
	IsEnabled() int32
	DisableService() int32
	EnableService() int32
	StartService() int32
	StopSevice() int32
	RestartService() int32
}

type service struct {
	*api.BaseClient
	timeouts ServiceTimeouts
	attempts ServiceAttempts
}

type ServiceInfo struct {
	Version string
	Status  Status
}

func (s *service) ListServices(ctx context.Context) ([]string, error) {
	thisCall, err := s.PrepareCall(
		s.timeouts.ListServices(),
		s.attempts.ListServices(),
		serviceRPC,
		serviceMethodNames,
	)
	if err != nil {
		return nil, err
	}
	return api.APICall(ctx, thisCall, stringSliceTransformer{})
}

func (s *service) IsEnabled(ctx context.Context, serviceName string) (bool, error) {
	thisCall, err := s.PrepareCall(
		s.timeouts.IsEnabled(),
		s.attempts.IsEnabled(),
		serviceRPC,
		serviceMethodEnabled,
		serviceName,
	)
	if err != nil {
		return false, err
	}
	return api.APICall(ctx, thisCall, booleanTransformer{})
}

func (s *service) DisableService(ctx context.Context, serviceName string) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.DisableService(),
		s.attempts.DisableService(),
		serviceRPC,
		serviceMethodDisable,
		serviceName,
	)
	if err != nil {
		return err
	}
	result, err := api.APICall(ctx, thisCall, booleanTransformer{})
	if err != nil {
		return err
	}

	if !result {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) EnableService(ctx context.Context, serviceName string) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.EnableService(),
		s.attempts.EnableService(),
		serviceRPC,
		serviceMethodEnable,
		serviceName,
	)
	if err != nil {
		return err
	}
	result, err := api.APICall(ctx, thisCall, booleanTransformer{})
	if err != nil {
		return err
	}

	if !result {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) StartService(ctx context.Context, serviceName string) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.StartService(),
		s.attempts.StartService(),
		serviceRPC,
		serviceMethodStart,
		serviceName,
	)
	if err != nil {
		return err
	}
	result, err := api.APICall(ctx, thisCall, booleanTransformer{})
	if err != nil {
		return err
	}

	if !result {
		return ErrExecutionFailure
	}
	return nil
}

func (s *service) StopSevice(ctx context.Context, serviceName string) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.StopSevice(),
		s.attempts.StopSevice(),
		serviceRPC,
		serviceMethodStop,
		serviceName,
	)
	if err != nil {
		return err
	}
	result, err := api.APICall(ctx, thisCall, booleanTransformer{})
	if err != nil {
		return err
	}

	if !result {
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
