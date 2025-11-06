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
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultGetAllTimeout         time.Duration = 30 * time.Second
	defaultTSetTimeout                         = 30 * time.Second
	defaultAddTimeout                          = 30 * time.Second
	defaultDeleteTimeout                       = 30 * time.Second
	defaultCommitOrRevertTimeout               = 30 * time.Second
)

type SystemTimeouts interface {
	GetAll() time.Duration
	TSet() time.Duration
	Add() time.Duration
	Delete() time.Duration
	CommitOrRevert() time.Duration
}

type SystemTimeoutsModel struct {
	GetAllTimeout         types.String `tfsdk:"get_all"`
	TSetTimeout           types.String `tfsdk:"t_set"`
	AddTimeout            types.String `tfsdk:"add"`
	DeleteTimeout         types.String `tfsdk:"delete"`
	CommitOrRevertTimeout types.String `tfsdk:"commit_or_revert"`
}

type SystemFacade interface {
	GetAll(ctx context.Context, section ...any) ([]System, error)
	GetSystem(ctx context.Context) (*System, error)
	TSet(ctx context.Context, data any, section ...any) error
	Add(ctx context.Context, section ...any) (string, error)
	Delete(ctx context.Context, section ...any) error
	CommitOrRevert(ctx context.Context, section ...any) error
}

type systemTimeouts struct {
	getAllTimeout,
	tSetTimeout,
	addTimeout,
	deleteTimeout,
	commitOrRevertTimeout time.Duration
}

func (sT *systemTimeouts) GetAll() time.Duration {
	return sT.getAllTimeout
}

func (sT *systemTimeouts) TSet() time.Duration {
	return sT.tSetTimeout
}

func (sT *systemTimeouts) Add() time.Duration {
	return sT.addTimeout
}

func (sT *systemTimeouts) Delete() time.Duration {
	return sT.deleteTimeout
}

func (sT *systemTimeouts) CommitOrRevert() time.Duration {
	return sT.commitOrRevertTimeout
}

var (
	_ SystemFacade   = (*system)(nil)
	_ WithSession    = (*system)(nil)
	_ SystemTimeouts = (*systemTimeouts)(nil)

	uciTimeoutSchemaAttribute = schema.SingleNestedAttribute{
		MarkdownDescription: `Uci operations timeout configuration`,
		Description:         `Uci operations timeout configuration`,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"get_all": schema.StringAttribute{
				MarkdownDescription: `Get all RPC timeout value`,
				Description:         `Get all RPC timeout value`,
				Optional:            true,
			},
			"t_set": schema.StringAttribute{
				MarkdownDescription: `T set RPC timeout value`,
				Description:         `T set RPC timeout value`,
				Optional:            true,
			},
			"add": schema.StringAttribute{
				MarkdownDescription: `Add RPC timeout value`,
				Description:         `Add RPC timeout value`,
				Optional:            true,
			},
			"delete": schema.StringAttribute{
				MarkdownDescription: `Delete RPC timeout value`,
				Description:         `Delete RPC timeout value`,
				Optional:            true,
			},
			"commit_or_revert": schema.StringAttribute{
				MarkdownDescription: `Commit or revert operation timeout configuration`,
				Description:         `Commit or revert operation timeout configuration`,
				Optional:            true,
			},
		},
	}
)

func parseSystemTimeouts(ctx context.Context, t *TimeoutsModel) (SystemTimeouts, error) {
	getAllTimeout := defaultGetAllTimeout
	tSetTimeout := defaultTSetTimeout
	addTimeout := defaultAddTimeout
	deleteTimeout := defaultDeleteTimeout
	commitOrRevertTimeout := defaultCommitOrRevertTimeout

	if t != nil && t.System != nil && !t.System.GetAllTimeout.IsNull() {
		parsedGetAllTimeout, err := time.ParseDuration(t.System.GetAllTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		getAllTimeout = parsedGetAllTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: get_all config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default get_all config")
	}

	if t != nil && t.System != nil && !t.System.TSetTimeout.IsNull() {
		parsedTSetTimeout, err := time.ParseDuration(t.System.TSetTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		tSetTimeout = parsedTSetTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: t_set config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default t_set config")
	}

	if t != nil && t.System != nil && !t.System.AddTimeout.IsNull() {
		parsedAddTimeout, err := time.ParseDuration(t.System.AddTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		addTimeout = parsedAddTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: add config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default add config")
	}

	if t != nil && t.System != nil && !t.System.DeleteTimeout.IsNull() {
		parsedDeleteTimeout, err := time.ParseDuration(t.System.DeleteTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		deleteTimeout = parsedDeleteTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: delete config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default delete config")
	}

	if t != nil && t.System != nil && !t.System.CommitOrRevertTimeout.IsNull() {
		parsedCommitOrRevertTimeout, err := time.ParseDuration(t.System.CommitOrRevertTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		commitOrRevertTimeout = parsedCommitOrRevertTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: commit_or_revert config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default commit_or_revert config")
	}

	return &systemTimeouts{
		getAllTimeout,
		tSetTimeout,
		addTimeout,
		deleteTimeout,
		commitOrRevertTimeout,
	}, nil
}

type system struct {
	timeouts SystemTimeouts

	token  string
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

func (c *system) SetToken(ctx context.Context, token string) error {
	c.token = token
	return nil
}

func (c *system) GetAll(ctx context.Context, sections ...any) ([]System, error) {
	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections specified")
	}

	result, err := call(ctx, c.client, c.timeouts.GetAll(),
		*c.url, c.token, "uci", "get_all", sections)
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

func (c *system) GetSystem(ctx context.Context) (*System, error) {
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

func (c *system) TSet(ctx context.Context, data any, section ...any) error {
	data, err := purgeFields(&data)
	if err != nil {
		return err
	}
	section = append(section, data)
	_, err = call(ctx, c.client, c.timeouts.TSet(),
		*c.url, c.token, "uci", "tset", section)
	return err
}

func (c *system) Add(ctx context.Context, section ...any) (string, error) {
	raw, err := call(ctx, c.client, c.timeouts.Add(),
		*c.url, c.token, "uci", "add", section)
	if err != nil {
		return "", err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", err
	}
	return s, nil
}

func (c *system) Delete(ctx context.Context, section ...any) error {
	_, err := call(ctx, c.client, c.timeouts.Delete(),
		*c.url, c.token, "uci", "delete", section)
	return err
}

func (c *system) uciCommit(ctx context.Context, section ...any) error {
	resp, err := call(ctx, c.client, c.timeouts.CommitOrRevert(),
		*c.url, c.token, "uci", "commit", section)
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

func (c *system) uciRevert(ctx context.Context, section ...any) error {
	resp, err := call(ctx, c.client, c.timeouts.CommitOrRevert(),
		*c.url, c.token, "uci", "revert", section)
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

func (c *system) CommitOrRevert(ctx context.Context, section ...any) error {
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
