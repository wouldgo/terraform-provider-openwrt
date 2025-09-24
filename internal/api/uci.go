// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
)

type SystemFacade interface {
	GetAll(ctx context.Context, section ...any) ([]System, error)
	GetSystem(ctx context.Context) (*System, error)
	TSet(ctx context.Context, data any, section ...any) error
	Add(ctx context.Context, section ...any) (string, error)
	Delete(ctx context.Context, section ...any) error
	CommitOrRevert(ctx context.Context, section ...any) error
}

var _ SystemFacade = (*uci)(nil)

type uci struct {
	token  *string
	url    *string
	client *http.Client
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

func (c *uci) GetAll(ctx context.Context, sections ...any) ([]System, error) {
	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections specified")
	}

	result, err := call(ctx, c.client, *c.url, *c.token, "uci", "get_all", sections)
	if err != nil {
		return nil, err
	}
	var data map[string]System
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data from the %v sections", sections)
	}

	return slices.Collect(maps.Values(data)), nil
}

func (c *uci) GetSystem(ctx context.Context) (*System, error) {
	result, err := c.GetAll(ctx, "system")
	if err != nil {
		return nil, err
	}

	for _, aResult := range result {
		if aResult.Anonymous && aResult.Type == "system" {
			return &aResult, nil
		}
	}

	return nil, fmt.Errorf("system section not found")
}

func (c *uci) TSet(ctx context.Context, data any, section ...any) error {
	data, err := purgeFields(&data)
	if err != nil {
		return err
	}
	section = append(section, data)
	_, err = call(ctx, c.client, *c.url, *c.token, "uci", "tset", section)
	return err
}

func (c *uci) Add(ctx context.Context, section ...any) (string, error) {
	raw, err := call(ctx, c.client, *c.url, *c.token, "uci", "add", section)
	if err != nil {
		return "", err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", err
	}
	return s, nil
}

func (c *uci) Delete(ctx context.Context, section ...any) error {
	_, err := call(ctx, c.client, *c.url, *c.token, "uci", "delete", section)
	return err
}

func (c *uci) uciCommit(ctx context.Context, section ...any) error {
	resp, err := call(ctx, c.client, *c.url, *c.token, "uci", "commit", section)
	if err != nil {
		return fmt.Errorf("uci commit call ko: %w", err)
	}

	var result bool
	if err = json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("error parsing json result from uci commit response: %w", err)
	}

	if !result {
		return fmt.Errorf("uci commit not ok")
	}
	return err
}

func (c *uci) uciRevert(ctx context.Context, section ...any) error {
	resp, err := call(ctx, c.client, *c.url, *c.token, "uci", "revert", section)
	if err != nil {
		return fmt.Errorf("uci revert call ko: %w", err)
	}

	var result bool
	if err = json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("error parsing json result from uci revert response: %w", err)
	}

	if !result {
		return fmt.Errorf("uci revert not ok")
	}
	return err
}

func (c *uci) CommitOrRevert(ctx context.Context, section ...any) error {
	toReturn := make([]error, 0, 2)
	err := c.uciCommit(ctx, section...)
	if err != nil {
		toReturn = append(toReturn, fmt.Errorf("failed to commit config %q: %w", section, err))
		err = c.uciRevert(ctx, section...)
		if err != nil {
			toReturn = append(toReturn, fmt.Errorf("failed to revert config %q: %w", section, err))
		}
	}

	if len(toReturn) > 0 {
		return errors.Join(toReturn...)
	}
	return nil
}
