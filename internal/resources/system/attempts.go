// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultGetAllAttempts         int32 = 1
	defaultTSetAttempts           int32 = 1
	defaultAddAttempts            int32 = 1
	defaultDeleteAttempts         int32 = 1
	defaultCommitOrRevertAttempts int32 = 1
)

var (
	_               rpc.SystemAttempts = systemAttempts{}
	DefaultAttempts                    = systemAttempts{
		defaultGetAllAttempts,
		defaultTSetAttempts,
		defaultAddAttempts,
		defaultDeleteAttempts,
		defaultCommitOrRevertAttempts,
	}
)

type systemAttempts struct {
	getAllAttempts,
	tSetAttempts,
	addAttempts,
	deleteAttempts,
	commitOrRevertAttempts int32
}

func (sR systemAttempts) GetAll() int32 {
	return sR.getAllAttempts
}

func (sR systemAttempts) TSet() int32 {
	return sR.tSetAttempts
}

func (sR systemAttempts) Add() int32 {
	return sR.addAttempts
}

func (sR systemAttempts) Delete() int32 {
	return sR.deleteAttempts
}

func (sR systemAttempts) CommitOrRevert() int32 {
	return sR.commitOrRevertAttempts
}

func ParseAttempts(ctx context.Context, t *SystemAttemptsModel) (rpc.SystemAttempts, error) {
	getAllAttempts := defaultGetAllAttempts
	tSetAttempts := defaultTSetAttempts
	addAttempts := defaultAddAttempts
	deleteAttempts := defaultDeleteAttempts
	commitOrRevertAttempts := defaultCommitOrRevertAttempts

	if t != nil && !t.GetAllAttempts.IsNull() {
		parsedGetAllAttempts := t.GetAllAttempts.ValueInt32()

		getAllAttempts = parsedGetAllAttempts
		tflog.Debug(ctx, "system - parse attempts configuration: get_all config parsed")
	} else {
		tflog.Debug(ctx, "system - parse attempts configuration: default get_all config")
	}

	if t != nil && !t.TSetAttempts.IsNull() {
		parsedTSetAttempts := t.TSetAttempts.ValueInt32()

		tSetAttempts = parsedTSetAttempts
		tflog.Debug(ctx, "system - parse attempts configuration: t_set config parsed")
	} else {
		tflog.Debug(ctx, "system - parse attempts configuration: default t_set config")
	}

	if t != nil && !t.AddAttempts.IsNull() {
		parsedAddAttempts := t.AddAttempts.ValueInt32()

		addAttempts = parsedAddAttempts
		tflog.Debug(ctx, "system - parse attempts configuration: add config parsed")
	} else {
		tflog.Debug(ctx, "system - parse attempts configuration: default add config")
	}

	if t != nil && !t.DeleteAttempts.IsNull() {
		parsedDeleteAttempts := t.DeleteAttempts.ValueInt32()

		deleteAttempts = parsedDeleteAttempts
		tflog.Debug(ctx, "system - parse attempts configuration: delete config parsed")
	} else {
		tflog.Debug(ctx, "system - parse attempts configuration: default delete config")
	}

	if t != nil && !t.CommitOrRevertAttempts.IsNull() {
		parsedCommitOrRevertAttempts := t.CommitOrRevertAttempts.ValueInt32()

		commitOrRevertAttempts = parsedCommitOrRevertAttempts
		tflog.Debug(ctx, "system - parse attempts configuration: commit_or_revert config parsed")
	} else {
		tflog.Debug(ctx, "system - parse attempts configuration: default commit_or_revert config")
	}

	toReturn := systemAttempts{
		getAllAttempts,
		tSetAttempts,
		addAttempts,
		deleteAttempts,
		commitOrRevertAttempts,
	}
	tflog.Debug(ctx, "system - attempts configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
