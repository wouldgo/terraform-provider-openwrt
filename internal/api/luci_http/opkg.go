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
	rawResult, err := c.Call(
		ctx,
		c.timeouts.UpdatePackages(),
		opkgfRPC,
		opkgMethodUpdate,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.OpkgErrCheckerTransformer.Transform(rawResult)
	return err
}

func (c *opkg) CheckPackage(ctx context.Context, pack string) (luci.PackageInfo, error) {
	var zero luci.PackageInfo
	rawResult, err := c.Call(
		ctx,
		c.timeouts.CheckPackage(),
		opkgfRPC,
		opkgMethodStatus,
		pack,
	)
	if err != nil {
		return zero, err
	}

	opkgPackagerTransformer := http_transformers.OpkgPackager{
		Package: pack,
	}
	return opkgPackagerTransformer.Transform(rawResult)
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

	rawResult, err := c.Call(
		ctx,
		c.timeouts.InstallPackages(),
		opkgfRPC,
		opkgMethodInstall,
		toApi...,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.OpkgErrCheckerTransformer.Transform(rawResult)
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

	rawResult, err := c.Call(
		ctx,
		c.timeouts.RemovePackages(),
		opkgfRPC,
		opkgMethodRemove,
		toApi...,
	)
	if err != nil {
		return err
	}

	_, err = http_transformers.OpkgErrCheckerTransformer.Transform(rawResult)
	return err
}
