// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package fs

import (
	"context"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultWriteFileAttempts  int32 = 1
	defaultReadFileAttempts   int32 = 1
	defaultRemoveFileAttempts int32 = 1
)

var (
	_               rpc.FsAttempts = fsAttempts{}
	DefaultAttempts                = fsAttempts{
		defaultWriteFileAttempts,
		defaultReadFileAttempts,
		defaultRemoveFileAttempts,
	}
)

type fsAttempts struct {
	writeFileAttempts, readFileAttempts, removeFileAttempts int32
}

func (fsR fsAttempts) WriteFile() int32 {
	return fsR.writeFileAttempts
}

func (fsR fsAttempts) ReadFile() int32 {
	return fsR.readFileAttempts
}

func (fsR fsAttempts) RemoveFile() int32 {
	return fsR.removeFileAttempts
}

func ParseAttempts(ctx context.Context, f *FsAttemptsModel) (rpc.FsAttempts, error) {
	writeFileAttempts := defaultWriteFileAttempts
	readFileAttempts := defaultReadFileAttempts
	removeFileAttempts := defaultRemoveFileAttempts

	if f != nil && !f.WriteFileAttempts.IsNull() {
		parsedWriteFileAttempts := f.WriteFileAttempts.ValueInt32()

		writeFileAttempts = parsedWriteFileAttempts
		tflog.Debug(ctx, "fs - parse attempts configuration: write_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse attempts configuration: default write_file config")
	}

	if f != nil && !f.ReadFileAttempts.IsNull() {
		parsedReadFileAttempts := f.ReadFileAttempts.ValueInt32()

		readFileAttempts = parsedReadFileAttempts
		tflog.Debug(ctx, "fs - parse attempts configuration: read_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse attempts configuration: default read_file config")
	}

	if f != nil && !f.RemoveFileAttempts.IsNull() {
		parsedRemoveFileAttempts := f.RemoveFileAttempts.ValueInt32()

		removeFileAttempts = parsedRemoveFileAttempts
		tflog.Debug(ctx, "fs - parse attempts configuration: remove_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse attempts configuration: default remove_file config")
	}

	toReturn := fsAttempts{
		writeFileAttempts,
		readFileAttempts,
		removeFileAttempts,
	}
	tflog.Debug(ctx, "fs - attempts configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
