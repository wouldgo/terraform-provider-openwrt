// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http

import (
	"context"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/http"
	http_transformers "github.com/foxboron/terraform-provider-openwrt/internal/http/transformers"
)

const (
	fsRPC          = "fs"
	fsMethodWrite  = "writefile"
	fsMethodRead   = "readfile"
	fsMethodRemove = "remove"
)

var (
	_ luci.FsFacade = (*fs)(nil)
)

type fs struct {
	*http.BaseClient
	timeouts luci.FsTimeouts
}

func (c *fs) Writefile(ctx context.Context, path string, data []byte) error {
	rawResult, err := c.Call(
		ctx,
		c.timeouts.WriteFile(),
		fsRPC,
		fsMethodWrite,
		path,
		data,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.IgnoreOutputTransformer.Transform(rawResult)
	return err
}

func (c *fs) ReadFile(ctx context.Context, path string) ([]byte, error) {
	rawResult, err := c.Call(
		ctx,
		c.timeouts.ReadFile(),
		fsRPC,
		fsMethodRead,
		path,
	)
	if err != nil {
		return nil, err
	}

	return http_transformers.Base64StringerTransformer.Transform(rawResult)
}

func (c *fs) RemoveFile(ctx context.Context, path string) error {
	rawResult, err := c.Call(
		ctx,
		c.timeouts.RemoveFile(),
		fsRPC,
		fsMethodRemove,
		path,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.IgnoreOutputTransformer.Transform(rawResult)
	return err
}
