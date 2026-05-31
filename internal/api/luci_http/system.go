// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/http"
	http_transformers "github.com/foxboron/terraform-provider-openwrt/internal/http/transformers"
)

const (
	uciRPC          = "uci"
	uciMethodGetAll = "get_all"
	uciMethodTSet   = "tset"
	uciMethodAdd    = "add"
	uciMethodDelete = "delete"
	uciMethodCommit = "commit"
	uciMethodRevert = "revert"
)

var (
	_ luci.SystemFacade = (*system)(nil)
)

type system struct {
	*http.BaseClient
	timeouts luci.SystemTimeouts
}

func (s *system) GetAll(ctx context.Context, sections ...any) ([]luci.System, error) {
	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections specified")
	}

	rawResult, err := s.Call(
		ctx,
		s.timeouts.GetAll(),
		uciRPC,
		uciMethodGetAll,
		sections...,
	)
	if err != nil {
		return nil, err
	}

	systemSliceTransformer := http_transformers.SystemSlice{
		Sections: sections,
	}
	return systemSliceTransformer.Transform(rawResult)
}

func (s *system) GetSystem(ctx context.Context) (luci.System, error) {
	var zero luci.System
	results, err := s.GetAll(ctx, "system")
	if err != nil {
		return zero, err
	}

	for _, aResult := range results {
		if aResult.Anonymous && aResult.Type == "system" {
			return aResult, nil
		}
	}

	return zero, fmt.Errorf("system section not found")
}

func (s *system) TSet(ctx context.Context, data any, section ...any) error {
	data, err := purgeFields(&data)
	if err != nil {
		return err
	}
	section = append(section, data)
	rawResult, err := s.Call(
		ctx,
		s.timeouts.TSet(),
		uciRPC,
		uciMethodTSet,
		section...,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.IgnoreOutputTransformer.Transform(rawResult)
	return err
}

func (s *system) Add(ctx context.Context, section ...any) (string, error) {
	var zero string
	rawResult, err := s.Call(
		ctx,
		s.timeouts.Add(),
		uciRPC,
		uciMethodAdd,
		section...,
	)
	if err != nil {
		return zero, err
	}

	return http_transformers.StringTransformer.Transform(rawResult)
}

func (s *system) Delete(ctx context.Context, section ...any) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.Delete(),
		uciRPC,
		uciMethodDelete,
		section...,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.IgnoreOutputTransformer.Transform(rawResult)
	return err
}

func (s *system) uciCommit(ctx context.Context, section ...any) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.CommitOrRevert(),
		uciRPC,
		uciMethodCommit,
		section...,
	)
	if err != nil {
		return err
	}

	result, err := http_transformers.BooleanTransformer.Transform(rawResult)
	if err != nil {
		return errors.Join(luci.ErrExecutionFailure, fmt.Errorf("uci commit call ko: %w", err))
	}

	if !result {
		return fmt.Errorf("uci commit not ok")
	}
	return err
}

func (s *system) uciRevert(ctx context.Context, section ...any) error {
	rawResult, err := s.Call(
		ctx,
		s.timeouts.CommitOrRevert(),
		uciRPC,
		uciMethodRevert,
		section...,
	)
	if err != nil {
		return err
	}

	result, err := http_transformers.BooleanTransformer.Transform(rawResult)
	if err != nil {
		return errors.Join(luci.ErrExecutionFailure, fmt.Errorf("uci commit call ko: %w", err))
	}

	if !result {
		return fmt.Errorf("uci revert not ok")
	}
	return err
}

func (s *system) CommitOrRevert(ctx context.Context, section ...any) error {
	toReturn := make([]error, 0, 2)
	err := s.uciCommit(ctx, section...)
	if err != nil {
		toReturn = append(toReturn, fmt.Errorf("failed to commit config %q: %w", section, err))
		err = s.uciRevert(ctx, section...)
		if err != nil {
			toReturn = append(toReturn, fmt.Errorf("failed to revert config %q: %w", section, err))
		}
	}

	if len(toReturn) > 0 {
		return errors.Join(toReturn...)
	}
	return nil
}

// Purge the sections from the anonymous things
func purgeFields(d any) (any, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(b, &objmap); err != nil {
		return nil, err
	}
	delete(objmap, ".name")
	delete(objmap, ".anonymous")
	delete(objmap, ".type")
	return objmap, nil
}
