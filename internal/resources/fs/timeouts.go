// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package fs

import (
	"context"
	"time"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultWriteFileTimeout  = 30 * time.Second
	defaultReadFileTimeout   = 30 * time.Second
	defaultRemoveFileTimeout = 30 * time.Second
)

var (
	_               rpc.FsTimeouts = fsTimeouts{}
	DefaultTimeouts                = fsTimeouts{
		defaultWriteFileTimeout,
		defaultReadFileTimeout,
		defaultRemoveFileTimeout,
	}
)

type fsTimeouts struct {
	writeFileTimeout, readFileTimeout, removeFileTimeout time.Duration
}

func (fsT fsTimeouts) WriteFile() time.Duration {
	return fsT.writeFileTimeout
}

func (fsT fsTimeouts) ReadFile() time.Duration {
	return fsT.readFileTimeout
}

func (fsT fsTimeouts) RemoveFile() time.Duration {
	return fsT.removeFileTimeout
}

func ParseTimeouts(ctx context.Context, t *FsTimeoutsModel) (rpc.FsTimeouts, error) {
	readFileTimeout := defaultReadFileTimeout
	removeFileTimeout := defaultRemoveFileTimeout
	writeFileTimeout := defaultWriteFileTimeout

	if t != nil && !t.ReadFileTimeout.IsNull() {
		parsedReadFileTimeout, err := time.ParseDuration(t.ReadFileTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		readFileTimeout = parsedReadFileTimeout
		tflog.Debug(ctx, "fs - parse timeout configuration: read_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse timeout configuration: default read_file config")
	}

	if t != nil && !t.RemoveFileTimeout.IsNull() {
		parsedRemoveFileTimeout, err := time.ParseDuration(t.RemoveFileTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		removeFileTimeout = parsedRemoveFileTimeout
		tflog.Debug(ctx, "fs - parse timeout configuration: remove_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse timeout configuration: default remove_file config")
	}

	if t != nil && !t.WriteFileTimeout.IsNull() {
		parsedWriteFileTimeout, err := time.ParseDuration(t.WriteFileTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		writeFileTimeout = parsedWriteFileTimeout
		tflog.Debug(ctx, "fs - parse timeout configuration: write_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse timeout configuration: default write_file config")
	}

	toReturn := fsTimeouts{
		writeFileTimeout,
		readFileTimeout,
		removeFileTimeout,
	}
	tflog.Debug(ctx, "fs - timeout configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
