// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"
)

type FsFacade interface {
	Writefile(ctx context.Context, path string, data []byte) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	RemoveFile(ctx context.Context, path string) error
}

type FsTimeouts interface {
	WriteFile() time.Duration
	ReadFile() time.Duration
	RemoveFile() time.Duration
}
