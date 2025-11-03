// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultWriteFileTimeout  time.Duration = 30 * time.Second
	defaultReadFileTimeout                 = 30 * time.Second
	defaultRemoveFileTimeout               = 30 * time.Second
)

type FsTimeouts interface {
	WriteFile() time.Duration
	ReadFile() time.Duration
	RemoveFile() time.Duration
}

type FsTimeoutsModel struct {
	WriteFileTimeout  types.String `tfsdk:"write_file"`
	ReadFileTimeout   types.String `tfsdk:"read_file"`
	RemoveFileTimeout types.String `tfsdk:"remove_file"`
}

type FsFacade interface {
	Writefile(ctx context.Context, path string, data []byte) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	RemoveFile(ctx context.Context, path string) error
}

type fsTimeouts struct {
	writeFileTimeout, readFileTimeout, removeFileTimeout time.Duration
}

func (fsT *fsTimeouts) WriteFile() time.Duration {
	return fsT.writeFileTimeout
}

func (fsT *fsTimeouts) ReadFile() time.Duration {
	return fsT.readFileTimeout
}

func (fsT *fsTimeouts) RemoveFile() time.Duration {
	return fsT.removeFileTimeout
}

var (
	_ FsFacade    = (*fs)(nil)
	_ WithSession = (*fs)(nil)
	_ FsTimeouts  = (*fsTimeouts)(nil)

	fsTimeoutSchemaAttribute = schema.SingleNestedAttribute{
		MarkdownDescription: `Filesystem operations timeout configuration`,
		Description:         `Filesystem operations timeout configuration`,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"write_file": schema.StringAttribute{
				MarkdownDescription: `Write file filesystem operation timeout configuration`,
				Description:         `Write file filesystem operation timeout configuration`,
				Optional:            true,
			},
			"read_file": schema.StringAttribute{
				MarkdownDescription: `Read file filesystem operation timeout configuration`,
				Description:         `Read file filesystem operation timeout configuration`,
				Optional:            true,
			},
			"remove_file": schema.StringAttribute{
				MarkdownDescription: `Remove file filesystem operation timeout configuration`,
				Description:         `Remove file filesystem operation timeout configuration`,
				Optional:            true,
			},
		},
	}
)

func parseFsTimeouts(ctx context.Context, t *TimeoutsModel) (FsTimeouts, error) {
	readFileTimeout := defaultReadFileTimeout
	removeFileTimeout := defaultRemoveFileTimeout
	writeFileTimeout := defaultWriteFileTimeout

	if t != nil && t.Fs != nil && !t.Fs.ReadFileTimeout.IsNull() {
		parsedReadFileTimeout, err := time.ParseDuration(t.Fs.ReadFileTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		readFileTimeout = parsedReadFileTimeout
		tflog.Debug(ctx, "fs - parse timeout configuration: read_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse timeout configuration: default read_file config")
	}

	if t != nil && t.Fs != nil && !t.Fs.RemoveFileTimeout.IsNull() {
		parsedRemoveFileTimeout, err := time.ParseDuration(t.Fs.RemoveFileTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		removeFileTimeout = parsedRemoveFileTimeout
		tflog.Debug(ctx, "fs - parse timeout configuration: remove_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse timeout configuration: default remove_file config")
	}

	if t != nil && t.Fs != nil && !t.Fs.WriteFileTimeout.IsNull() {
		parsedWriteFileTimeout, err := time.ParseDuration(t.Fs.WriteFileTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		writeFileTimeout = parsedWriteFileTimeout
		tflog.Debug(ctx, "fs - parse timeout configuration: write_file config parsed")
	} else {
		tflog.Debug(ctx, "fs - parse timeout configuration: default write_file config")
	}

	toReturn := &fsTimeouts{
		writeFileTimeout,
		readFileTimeout,
		removeFileTimeout,
	}

	tflog.Debug(ctx, "fs - timeout configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})

	return toReturn, nil
}

type fs struct {
	timeouts FsTimeouts

	token  string
	url    *string
	client *http.Client
}

func (c *fs) SetToken(ctx context.Context, token string) error {
	c.token = token
	return nil
}

func (c *fs) Writefile(ctx context.Context, path string, data []byte) error {
	_, err := call(ctx, c.client, c.timeouts.WriteFile(),
		*c.url, c.token, "fs", "writefile", []any{path, data})
	return err
}

func (c *fs) ReadFile(ctx context.Context, path string) ([]byte, error) {
	raw, err := call(ctx, c.client, c.timeouts.ReadFile(),
		*c.url, c.token, "fs", "readfile", []any{path})
	if err != nil {
		return nil, err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

func (c *fs) RemoveFile(ctx context.Context, path string) error {
	_, err := call(ctx, c.client, c.timeouts.RemoveFile(),
		*c.url, c.token, "fs", "remove", []any{path})
	return err
}
