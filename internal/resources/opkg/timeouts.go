// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package opkg

import (
	"context"
	"time"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultUpdatePackagesTimeout  = 30 * time.Second
	defaultCheckPackageTimeout    = 30 * time.Second
	defaultInstallPackagesTimeout = 30 * time.Second
	defaultRemovePackagesTimeout  = 30 * time.Second
)

var (
	_               rpc.OpkgTimeouts = opkgTimeouts{}
	DefaultTimeouts                  = opkgTimeouts{
		defaultUpdatePackagesTimeout,
		defaultCheckPackageTimeout,
		defaultInstallPackagesTimeout,
		defaultRemovePackagesTimeout,
	}
)

type opkgTimeouts struct {
	updatePackagesTimeout,
	checkPackageTimeout,
	installPackagesTimeout,
	removePackagesTimeout time.Duration
}

func (opkgT opkgTimeouts) UpdatePackages() time.Duration {
	return opkgT.updatePackagesTimeout
}

func (opkgT opkgTimeouts) CheckPackage() time.Duration {
	return opkgT.checkPackageTimeout
}

func (opkgT opkgTimeouts) InstallPackages() time.Duration {
	return opkgT.installPackagesTimeout
}

func (opkgT opkgTimeouts) RemovePackages() time.Duration {
	return opkgT.removePackagesTimeout
}

func ParseTimeouts(ctx context.Context, t *OpkgTimeoutsModel) (rpc.OpkgTimeouts, error) {
	updatePackagesTimeout := defaultUpdatePackagesTimeout
	checkPackageTimeout := defaultCheckPackageTimeout
	installPackagesTimeout := defaultInstallPackagesTimeout
	removePackagesTimeout := defaultRemovePackagesTimeout

	if t != nil && !t.UpdatePackagesTimeout.IsNull() {
		parsedUpdatePackagesTimeout, err := time.ParseDuration(t.UpdatePackagesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		updatePackagesTimeout = parsedUpdatePackagesTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: update_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default update_packages config")
	}

	if t != nil && !t.CheckPackageTimeout.IsNull() {
		parsedCheckPackageTimeout, err := time.ParseDuration(t.CheckPackageTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		checkPackageTimeout = parsedCheckPackageTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: check_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default check_packages config")
	}

	if t != nil && !t.InstallPackagesTimeout.IsNull() {
		parsedInstallPackagesTimeout, err := time.ParseDuration(t.InstallPackagesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		installPackagesTimeout = parsedInstallPackagesTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: install_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default install_packages config")
	}

	if t != nil && !t.RemovePackagesTimeout.IsNull() {
		parsedRemovePackagesTimeout, err := time.ParseDuration(t.RemovePackagesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		removePackagesTimeout = parsedRemovePackagesTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: remove_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default remove_packages config")
	}

	toReturn := opkgTimeouts{
		updatePackagesTimeout,
		checkPackageTimeout,
		installPackagesTimeout,
		removePackagesTimeout,
	}
	tflog.Debug(ctx, "opkg - timeout configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
