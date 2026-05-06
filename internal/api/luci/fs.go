// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
)

const (
	fsRPC          = "fs"
	fsMethodWrite  = "writefile"
	fsMethodRead   = "readfile"
	fsMethodRemove = "remove"
)

type FsFacade interface {
	Writefile(ctx context.Context, path string, data []byte) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	RemoveFile(ctx context.Context, path string) error
}

var (
	_ FsFacade = (*fs)(nil)
)

type FsTimeouts interface {
	WriteFile() time.Duration
	ReadFile() time.Duration
	RemoveFile() time.Duration
}

type fs struct {
	*api.BaseClient
	timeouts FsTimeouts
}

func (c *fs) Writefile(ctx context.Context, path string, data []byte) error {
	thisCall, err := c.PrepareCall(
		c.timeouts.WriteFile(),
		fsRPC,
		fsMethodWrite,
		path,
		data,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(
		ctx,
		thisCall,
		ignoreOutput{},
	)
	return err
}

func (c *fs) ReadFile(ctx context.Context, path string) ([]byte, error) {
	thisCall, err := c.PrepareCall(
		c.timeouts.ReadFile(),
		fsRPC,
		fsMethodRead,
		path,
	)
	if err != nil {
		return nil, err
	}
	return api.APICall(
		ctx,
		thisCall,
		base64Stringer{},
	)
}

func (c *fs) RemoveFile(ctx context.Context, path string) error {
	thisCall, err := c.PrepareCall(
		c.timeouts.RemoveFile(),
		fsRPC,
		fsMethodRemove,
		path,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(
		ctx,
		thisCall,
		ignoreOutput{},
	)
	return err
}
