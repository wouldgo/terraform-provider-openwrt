// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"
)

type OpkgFacade interface {
	UpdatePackages(ctx context.Context) error
	CheckPackage(ctx context.Context, pack string) (PackageInfo, error)
	InstallPackages(ctx context.Context, packages ...string) error
	RemovePackages(ctx context.Context, packages ...string) error
}

type OpkgTimeouts interface {
	UpdatePackages() time.Duration
	CheckPackage() time.Duration
	InstallPackages() time.Duration
	RemovePackages() time.Duration
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
