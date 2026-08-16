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
	_, err := http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.WriteFile(),
		fsRPC,
		fsMethodWrite,
		http_transformers.IgnoreOutputTransformer,
		path,
		data,
	)

	return err
}

func (c *fs) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.ReadFile(),
		fsRPC,
		fsMethodRead,
		http_transformers.Base64StringerTransformer,
		path,
	)
}

func (c *fs) RemoveFile(ctx context.Context, path string) error {
	_, err := http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.RemoveFile(),
		fsRPC,
		fsMethodRemove,
		http_transformers.IgnoreOutputTransformer,
		path,
	)

	return err
}
