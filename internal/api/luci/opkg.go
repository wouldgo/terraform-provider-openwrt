// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
)

const (
	opkgfRPC          = "ipkg"
	opkgMethodUpdate  = "update"
	opkgMethodStatus  = "status"
	opkgMethodInstall = "install"
	opkgMethodRemove  = "remove"
)

type OpkgFacade interface {
	UpdatePackages(ctx context.Context) error
	CheckPackage(ctx context.Context, pack string) (PackageInfo, error)
	InstallPackages(ctx context.Context, packages ...string) error
	RemovePackages(ctx context.Context, packages ...string) error
}

var (
	_ OpkgFacade = (*opkg)(nil)
)

type OpkgTimeouts interface {
	UpdatePackages() time.Duration
	CheckPackage() time.Duration
	InstallPackages() time.Duration
	RemovePackages() time.Duration
}

type OpkgAttempts interface {
	UpdatePackages() int32
	CheckPackage() int32
	InstallPackages() int32
	RemovePackages() int32
}

type opkg struct {
	*api.BaseClient
	timeouts OpkgTimeouts
	attempts OpkgAttempts
}

type PackageInfo struct {
	Version string `json:"Version"`
	Status  Status `json:"Status"`
}

type Status struct {
	Installed bool `json:"installed"`
	// User      bool `json:"user"`
	// Install   bool `json:"install"`
}

func (c *opkg) UpdatePackages(ctx context.Context) error {
	thisCall, err := c.PrepareCall(
		c.timeouts.UpdatePackages(),
		c.attempts.UpdatePackages(),
		opkgfRPC,
		opkgMethodUpdate,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(ctx, thisCall, opkgErrChecker{})
	return err
}

func (c *opkg) CheckPackage(ctx context.Context, pack string) (PackageInfo, error) {
	var zero PackageInfo
	thisCall, err := c.PrepareCall(
		c.timeouts.CheckPackage(),
		c.attempts.CheckPackage(),
		opkgfRPC,
		opkgMethodStatus,
		pack,
	)
	if err != nil {
		return zero, err
	}
	return api.APICall(ctx, thisCall, opkgPackager{
		pack: pack,
	})
}

func (c *opkg) InstallPackages(ctx context.Context, packages ...string) error {
	packagesLen := len(packages)
	if packagesLen == 0 {
		return ErrPackagesNotSpecified
	}

	toApi := make([]any, 0, packagesLen)
	for _, aPackage := range packages {
		toApi = append(toApi, aPackage)
	}

	thisCall, err := c.PrepareCall(
		c.timeouts.InstallPackages(),
		c.attempts.InstallPackages(),
		opkgfRPC,
		opkgMethodInstall,
		toApi...,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(ctx, thisCall, opkgErrChecker{})
	return err
}

func (c *opkg) RemovePackages(ctx context.Context, packages ...string) error {
	packagesLen := len(packages)
	if packagesLen == 0 {
		return ErrPackageNotFound
	}

	toApi := make([]any, 0, packagesLen)
	for _, aPackage := range packages {
		toApi = append(toApi, aPackage)
	}

	thisCall, err := c.PrepareCall(
		c.timeouts.RemovePackages(),
		c.attempts.RemovePackages(),
		opkgfRPC,
		opkgMethodRemove,
		toApi...,
	)
	if err != nil {
		return err
	}
	_, err = api.APICall(ctx, thisCall, opkgErrChecker{})
	return err
}
