// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"time"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultGetAllTimeout         = 30 * time.Second
	defaultTSetTimeout           = 30 * time.Second
	defaultAddTimeout            = 30 * time.Second
	defaultDeleteTimeout         = 30 * time.Second
	defaultCommitOrRevertTimeout = 30 * time.Second
)

var (
	_               rpc.SystemTimeouts = systemTimeouts{}
	DefaultTimeouts                    = systemTimeouts{
		defaultGetAllTimeout,
		defaultTSetTimeout,
		defaultAddTimeout,
		defaultDeleteTimeout,
		defaultCommitOrRevertTimeout,
	}
)

type systemTimeouts struct {
	getAllTimeout,
	tSetTimeout,
	addTimeout,
	deleteTimeout,
	commitOrRevertTimeout time.Duration
}

func (sT systemTimeouts) GetAll() time.Duration {
	return sT.getAllTimeout
}

func (sT systemTimeouts) TSet() time.Duration {
	return sT.tSetTimeout
}

func (sT systemTimeouts) Add() time.Duration {
	return sT.addTimeout
}

func (sT systemTimeouts) Delete() time.Duration {
	return sT.deleteTimeout
}

func (sT systemTimeouts) CommitOrRevert() time.Duration {
	return sT.commitOrRevertTimeout
}

func ParseTimeouts(ctx context.Context, t *SystemTimeoutsModel) (rpc.SystemTimeouts, error) {
	getAllTimeout := defaultGetAllTimeout
	tSetTimeout := defaultTSetTimeout
	addTimeout := defaultAddTimeout
	deleteTimeout := defaultDeleteTimeout
	commitOrRevertTimeout := defaultCommitOrRevertTimeout

	if t != nil && !t.GetAllTimeout.IsNull() {
		parsedGetAllTimeout, err := time.ParseDuration(t.GetAllTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		getAllTimeout = parsedGetAllTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: get_all config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default get_all config")
	}

	if t != nil && !t.TSetTimeout.IsNull() {
		parsedTSetTimeout, err := time.ParseDuration(t.TSetTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		tSetTimeout = parsedTSetTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: t_set config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default t_set config")
	}

	if t != nil && !t.AddTimeout.IsNull() {
		parsedAddTimeout, err := time.ParseDuration(t.AddTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		addTimeout = parsedAddTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: add config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default add config")
	}

	if t != nil && !t.DeleteTimeout.IsNull() {
		parsedDeleteTimeout, err := time.ParseDuration(t.DeleteTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		deleteTimeout = parsedDeleteTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: delete config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default delete config")
	}

	if t != nil && !t.CommitOrRevertTimeout.IsNull() {
		parsedCommitOrRevertTimeout, err := time.ParseDuration(t.CommitOrRevertTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		commitOrRevertTimeout = parsedCommitOrRevertTimeout
		tflog.Debug(ctx, "system - parse timeout configuration: commit_or_revert config parsed")
	} else {
		tflog.Debug(ctx, "system - parse timeout configuration: default commit_or_revert config")
	}

	toReturn := systemTimeouts{
		getAllTimeout,
		tSetTimeout,
		addTimeout,
		deleteTimeout,
		commitOrRevertTimeout,
	}
	tflog.Debug(ctx, "system - timeout configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
