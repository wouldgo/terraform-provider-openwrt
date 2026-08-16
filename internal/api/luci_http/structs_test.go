// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http_test

import (
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
)

var (
	_ luci.Timeouts = mockTimeouts{}
)

type mockTimeouts struct {
	auth,
	writeFile,
	readFile,
	removeFile,
	updatePackages,
	checkPackage,
	installPackages,
	removePackages,
	listServices,
	isEnabled,
	disableService,
	enableService,
	startService,
	stopSevice,
	restartService,
	getAll,
	tSet,
	add,
	delete,
	commitOrRevert time.Duration
}

func (m mockTimeouts) Auth() time.Duration {
	return m.auth
}

func (m mockTimeouts) WriteFile() time.Duration {
	return m.writeFile
}
func (m mockTimeouts) ReadFile() time.Duration {
	return m.readFile
}
func (m mockTimeouts) RemoveFile() time.Duration {
	return m.removeFile
}

func (m mockTimeouts) UpdatePackages() time.Duration {
	return m.updatePackages
}
func (m mockTimeouts) CheckPackage() time.Duration {
	return m.checkPackage
}
func (m mockTimeouts) InstallPackages() time.Duration {
	return m.installPackages
}
func (m mockTimeouts) RemovePackages() time.Duration {
	return m.removePackages
}

func (m mockTimeouts) ListServices() time.Duration {
	return m.listServices
}
func (m mockTimeouts) IsEnabled() time.Duration {
	return m.isEnabled
}
func (m mockTimeouts) DisableService() time.Duration {
	return m.disableService
}
func (m mockTimeouts) EnableService() time.Duration {
	return m.enableService
}
func (m mockTimeouts) StartService() time.Duration {
	return m.startService
}
func (m mockTimeouts) StopService() time.Duration {
	return m.stopSevice
}
func (m mockTimeouts) RestartService() time.Duration {
	return m.restartService
}

func (m mockTimeouts) GetAll() time.Duration {
	return m.getAll
}
func (m mockTimeouts) TSet() time.Duration {
	return m.tSet
}
func (m mockTimeouts) Add() time.Duration {
	return m.add
}
func (m mockTimeouts) Delete() time.Duration {
	return m.delete
}
func (m mockTimeouts) CommitOrRevert() time.Duration {
	return m.commitOrRevert
}

func newMockTimeouts(duration time.Duration) luci.Timeouts {
	return mockTimeouts{
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
		duration,
	}
}
