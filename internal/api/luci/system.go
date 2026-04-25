// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
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

type SystemFacade interface {
	GetAll(ctx context.Context, section ...any) ([]System, error)
	GetSystem(ctx context.Context) (System, error)
	TSet(ctx context.Context, data any, section ...any) error
	Add(ctx context.Context, section ...any) (string, error)
	Delete(ctx context.Context, section ...any) error
	CommitOrRevert(ctx context.Context, section ...any) error
}

var (
	_ SystemFacade = (*system)(nil)
)

type SystemTimeouts interface {
	GetAll() time.Duration
	TSet() time.Duration
	Add() time.Duration
	Delete() time.Duration
	CommitOrRevert() time.Duration
}

type SystemAttempts interface {
	GetAll() int32
	TSet() int32
	Add() int32
	Delete() int32
	CommitOrRevert() int32
}

type system struct {
	*api.BaseClient
	timeouts SystemTimeouts
	attempts SystemAttempts
}

type System struct {
	Id        string `json:".name,omitempty"`
	Type      string `json:".type,omitzero,omitempty"`
	Anonymous bool   `json:".anonymous,omitzero,omitempty"`

	Hostname        string `json:"hostname,omitzero"`
	Description     string `json:"description,omitzero"`
	Notes           string `json:"notes,omitzero"`
	Buffersize      string `json:"buffersize,omitzero"`
	ConLogLevel     string `json:"conloglevel,omitzero"`
	CronLogLevel    string `json:"cronloglevel,omitzero"`
	KlogconLogLevel string `json:"klogconloglevel,omitzero"`
	LogBufferSize   string `json:"log_buffer_size,omitzero"`
	LogFile         string `json:"log_file,omitzero"`
	LogHostname     string `json:"log_hostname,omitzero"`
	LogIP           string `json:"log_ip,omitzero"`
	LogPort         string `json:"log_port,omitzero"`
	LogPrefix       string `json:"log_prefix,omitzero"`
	LogProto        string `json:"log_proto,omitzero"`
	LogRemote       string `json:"log_remote,omitzero"`
	LogSize         string `json:"log_size,omitzero"`
	LogTrailerNull  string `json:"log_trailer_null,omitzero"`
	LogType         string `json:"log_type,omitzero"`
	TTYLogin        string `json:"ttylogin,omitzero"`
	UrandomSeed     string `json:"urandom_seed,omitempty"`
	Timezone        string `json:"timezone,omitzero"`
	ZoneName        string `json:"zonename,omitzero"`
	ZramCompAlgo    string `json:"zram_comp_algo,omitzero"`
	ZramSizeMb      string `json:"zram_size_mb,omitzero"`
}

func (s *system) GetAll(ctx context.Context, sections ...any) ([]System, error) {
	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections specified")
	}

	thisCall, err := s.PrepareCall(
		s.timeouts.GetAll(),
		s.attempts.GetAll(),
		uciRPC,
		uciMethodGetAll,
		sections...,
	)
	if err != nil {
		return nil, err
	}
	return api.APICall(ctx, thisCall, systemSliceTransformer{
		sections: sections,
	})
}

func (s *system) GetSystem(ctx context.Context) (System, error) {
	var zero System
	result, err := s.GetAll(ctx, "system")
	if err != nil {
		return zero, err
	}

	for _, aResult := range result {
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
	thisCall, err := s.PrepareCall(
		s.timeouts.TSet(),
		s.attempts.TSet(),
		uciRPC,
		uciMethodTSet,
		section...,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(ctx, thisCall, ignoreOutput{})
	return err
}

func (s *system) Add(ctx context.Context, section ...any) (string, error) {
	var zero string
	thisCall, err := s.PrepareCall(
		s.timeouts.Add(),
		s.attempts.Add(),
		uciRPC,
		uciMethodAdd,
		section...,
	)
	if err != nil {
		return zero, err
	}
	return api.APICall(ctx, thisCall, stringTransformer{})
}

func (s *system) Delete(ctx context.Context, section ...any) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.Delete(),
		s.attempts.Delete(),
		uciRPC,
		uciMethodDelete,
		section...,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(ctx, thisCall, ignoreOutput{})
	return err
}

func (s *system) uciCommit(ctx context.Context, section ...any) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.CommitOrRevert(),
		s.attempts.CommitOrRevert(),
		uciRPC,
		uciMethodCommit,
		section...,
	)
	if err != nil {
		return err
	}
	result, err := api.APICall(ctx, thisCall, booleanTransformer{})
	if err != nil {
		return fmt.Errorf("uci commit call ko: %w", err)
	}

	if !result {
		return fmt.Errorf("uci commit not ok")
	}
	return err
}

func (s *system) uciRevert(ctx context.Context, section ...any) error {
	thisCall, err := s.PrepareCall(
		s.timeouts.CommitOrRevert(),
		s.attempts.CommitOrRevert(),
		uciRPC,
		uciMethodRevert,
		section...,
	)
	if err != nil {
		return err
	}
	result, err := api.APICall(ctx, thisCall, booleanTransformer{})
	if err != nil {
		return fmt.Errorf("uci commit call ko: %w", err)
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
