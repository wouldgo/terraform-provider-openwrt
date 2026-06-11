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
		return nil, luci.ErrSectionsNotSpecified
	}

	return http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.GetAll(),
		uciRPC,
		uciMethodGetAll,
		http_transformers.SystemSlice{
			Sections: sections,
		},
		sections...,
	)
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

	return zero, errors.Join(luci.ErrSectionNotFound, errors.New("system section"))
}

func (s *system) TSet(ctx context.Context, data any, sections ...any) error {
	if len(sections) == 0 {
		return luci.ErrSectionsNotSpecified
	}
	data, err := purgeFields(&data)
	if err != nil {
		return err
	}
	sections = append(sections, data)
	_, err = http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.TSet(),
		uciRPC,
		uciMethodTSet,
		http_transformers.IgnoreOutputTransformer,
		sections...,
	)

	return err
}

func (s *system) Add(ctx context.Context, sections ...any) (string, error) {
	var zero string
	if len(sections) == 0 {
		return zero, luci.ErrSectionsNotSpecified
	}
	return http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.Add(),
		uciRPC,
		uciMethodAdd,
		http_transformers.StringTransformer,
		sections...,
	)
}

func (s *system) Delete(ctx context.Context, sections ...any) error {
	if len(sections) == 0 {
		return luci.ErrSectionsNotSpecified
	}
	_, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.Delete(),
		uciRPC,
		uciMethodDelete,
		http_transformers.IgnoreOutputTransformer,
		sections...,
	)

	return err
}

func (s *system) uciCommit(ctx context.Context, sections ...any) error {
	if len(sections) == 0 {
		return luci.ErrSectionsNotSpecified
	}
	result, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.CommitOrRevert(),
		uciRPC,
		uciMethodCommit,
		http_transformers.BooleanTransformer,
		sections...,
	)
	if err != nil {
		return errors.Join(luci.ErrExecutionFailure, fmt.Errorf("uci commit call ko: %w", err))
	}

	if !result {
		return luci.ErrUCICommit
	}
	return err
}

func (s *system) uciRevert(ctx context.Context, sections ...any) error {
	if len(sections) == 0 {
		return luci.ErrSectionsNotSpecified
	}
	result, err := http.Call(
		ctx,
		s.BaseClient,
		s.timeouts.CommitOrRevert(),
		uciRPC,
		uciMethodRevert,
		http_transformers.BooleanTransformer,
		sections...,
	)
	if err != nil {
		return errors.Join(luci.ErrExecutionFailure, fmt.Errorf("uci revert call ko: %w", err))
	}

	if !result {
		return luci.ErrUCIRevert
	}
	return err
}

func (s *system) CommitOrRevert(ctx context.Context, sections ...any) error {
	toReturn := make([]error, 0, 2)
	err := s.uciCommit(ctx, sections...)
	if err != nil {
		toReturn = append(toReturn, fmt.Errorf("failed to commit config %q: %w", sections, err))
		err = s.uciRevert(ctx, sections...)
		if err != nil {
			toReturn = append(toReturn, fmt.Errorf("failed to revert config %q: %w", sections, err))
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
