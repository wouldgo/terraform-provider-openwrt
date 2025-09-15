// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type service struct {
	token  *string
	url    *string
	client *http.Client
}

type ServiceInfo struct {
	Version string
	Status  Status
}

func (s *service) ListServices(ctx context.Context) ([]string, error) {
	result, err := call(
		ctx, s.client, *s.url, *s.token,
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
	result, err := call(
		ctx, s.client, *s.url, *s.token,
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
	result, err := call(
		ctx, s.client, *s.url, *s.token,
		"sys", "init.disable", []any{serviceName})
	if err != nil {
		return err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	if !data {
		return errors.Join(ErrExecutionFailure)
	}
	return nil
}

func (s *service) EnableService(ctx context.Context, serviceName string) error {
	result, err := call(
		ctx, s.client, *s.url, *s.token,
		"sys", "init.enable", []any{serviceName})
	if err != nil {
		return err
	}

	var data bool
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	if !data {
		return errors.Join(ErrExecutionFailure)
	}
	return nil
}

func (s *service) StartService(ctx context.Context, serviceName string) error {
	result, err := call(
		ctx, s.client, *s.url, *s.token,
		"sys", "init.start", []any{serviceName})
	if err != nil {
		return err
	}

	var data any
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}
	fmt.Printf("data replied: %v", data)
	return nil
}

func (s *service) StopSevice(ctx context.Context, serviceName string) error {
	result, err := call(
		ctx, s.client, *s.url, *s.token,
		"sys", "init.stop", []any{serviceName})
	if err != nil {
		return err
	}

	var data any
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}
	fmt.Printf("data replied: %v", data)
	return nil
}

func (s *service) RestartService(ctx context.Context, serviceName string) error {
	err := s.StopSevice(ctx, serviceName)
	if err != nil {
		return err
	}
	return s.StartService(ctx, serviceName)
}
