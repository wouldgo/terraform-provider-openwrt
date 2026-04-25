// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package opkg

import (
	"context"

	rpc "github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultUpdatePackagesAttempts  int32 = 1
	defaultCheckPackageAttempts    int32 = 1
	defaultInstallPackagesAttempts int32 = 1
	defaultRemovePackagesAttempts  int32 = 1
)

var (
	_               rpc.OpkgAttempts = opkgAttempts{}
	DefaultAttempts                  = opkgAttempts{
		defaultUpdatePackagesAttempts,
		defaultCheckPackageAttempts,
		defaultInstallPackagesAttempts,
		defaultRemovePackagesAttempts,
	}
)

type opkgAttempts struct {
	updatePackagesAttempts,
	checkPackageAttempts,
	installPackagesAttempts,
	removePackagesAttempts int32
}

func (opkgR opkgAttempts) UpdatePackages() int32 {
	return opkgR.updatePackagesAttempts
}

func (opkgR opkgAttempts) CheckPackage() int32 {
	return opkgR.checkPackageAttempts
}

func (opkgR opkgAttempts) InstallPackages() int32 {
	return opkgR.installPackagesAttempts
}

func (opkgR opkgAttempts) RemovePackages() int32 {
	return opkgR.removePackagesAttempts
}

func ParseAttempts(ctx context.Context, t *OpkgAttemptsModel) (rpc.OpkgAttempts, error) {
	updatePackagesAttempts := defaultUpdatePackagesAttempts
	checkPackageAttempts := defaultCheckPackageAttempts
	installPackagesAttempts := defaultInstallPackagesAttempts
	removePackagesAttempts := defaultRemovePackagesAttempts

	if t != nil && !t.UpdatePackagesAttempts.IsNull() {
		parsedUpdatePackagesAttempts := t.UpdatePackagesAttempts.ValueInt32()

		updatePackagesAttempts = parsedUpdatePackagesAttempts
		tflog.Debug(ctx, "opkg - parse attempts configuration: update_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse attempts configuration: default update_packages config")
	}

	if t != nil && !t.CheckPackageAttempts.IsNull() {
		parsedCheckPackageAttempts := t.CheckPackageAttempts.ValueInt32()

		checkPackageAttempts = parsedCheckPackageAttempts
		tflog.Debug(ctx, "opkg - parse attempts configuration: check_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse attempts configuration: default check_packages config")
	}

	if t != nil && !t.InstallPackagesAttempts.IsNull() {
		parsedInstallPackagesAttempts := t.InstallPackagesAttempts.ValueInt32()

		installPackagesAttempts = parsedInstallPackagesAttempts
		tflog.Debug(ctx, "opkg - parse attempts configuration: install_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse attempts configuration: default install_packages config")
	}

	if t != nil && !t.RemovePackagesAttempts.IsNull() {
		parsedRemovePackagesAttempts := t.RemovePackagesAttempts.ValueInt32()

		removePackagesAttempts = parsedRemovePackagesAttempts
		tflog.Debug(ctx, "opkg - parse attempts configuration: remove_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse attempts configuration: default remove_packages config")
	}

	toReturn := opkgAttempts{
		updatePackagesAttempts,
		checkPackageAttempts,
		installPackagesAttempts,
		removePackagesAttempts,
	}
	tflog.Debug(ctx, "opkg - attempts configuration parsed", map[string]interface{}{
		"configuration": toReturn,
	})
	return toReturn, nil
}
