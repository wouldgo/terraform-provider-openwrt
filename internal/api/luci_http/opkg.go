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
	opkgfRPC          = "ipkg"
	opkgMethodUpdate  = "update"
	opkgMethodStatus  = "status"
	opkgMethodInstall = "install"
	opkgMethodRemove  = "remove"
)

var (
	_ luci.OpkgFacade = (*opkg)(nil)
)

type opkg struct {
	*http.BaseClient
	timeouts luci.OpkgTimeouts
}

func (c *opkg) UpdatePackages(ctx context.Context) error {
	_, err := http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.UpdatePackages(),
		opkgfRPC,
		opkgMethodUpdate,
		http_transformers.OpkgErrCheckerTransformer,
	)

	return err
}

func (c *opkg) CheckPackage(ctx context.Context, pack string) (luci.PackageInfo, error) {
	return http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.CheckPackage(),
		opkgfRPC,
		opkgMethodStatus,
		http_transformers.OpkgPackager{
			Package: pack,
		},
		pack,
	)
}

func (c *opkg) InstallPackages(ctx context.Context, packages ...string) error {
	packagesLen := len(packages)
	if packagesLen == 0 {
		return luci.ErrPackagesNotSpecified
	}

	toApi := make([]any, 0, packagesLen)
	for _, aPackage := range packages {
		toApi = append(toApi, aPackage)
	}

	_, err := http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.InstallPackages(),
		opkgfRPC,
		opkgMethodInstall,
		http_transformers.OpkgErrCheckerTransformer,
		toApi...,
	)
	if err != nil {
		return err
	}

	return err
}

func (c *opkg) RemovePackages(ctx context.Context, packages ...string) error {
	packagesLen := len(packages)
	if packagesLen == 0 {
		return luci.ErrPackageNotFound
	}

	toApi := make([]any, 0, packagesLen)
	for _, aPackage := range packages {
		toApi = append(toApi, aPackage)
	}

	_, err := http.Call(
		ctx,
		c.BaseClient,
		c.timeouts.RemovePackages(),
		opkgfRPC,
		opkgMethodRemove,
		http_transformers.OpkgErrCheckerTransformer,
		toApi...,
	)

	return err
}
