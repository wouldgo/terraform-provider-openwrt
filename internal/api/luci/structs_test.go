// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_test

import (
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
)

var (
	_ luci.Timeouts = mockTimeouts{}
	_ luci.Attempts = mockAttempts{}
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
func (m mockTimeouts) StopSevice() time.Duration {
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

type mockAttempts struct {
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
	commitOrRevert int32
}

func (m mockAttempts) WriteFile() int32 {
	return m.writeFile
}
func (m mockAttempts) ReadFile() int32 {
	return m.readFile
}
func (m mockAttempts) RemoveFile() int32 {
	return m.removeFile
}

func (m mockAttempts) UpdatePackages() int32 {
	return m.updatePackages
}
func (m mockAttempts) CheckPackage() int32 {
	return m.checkPackage
}
func (m mockAttempts) InstallPackages() int32 {
	return m.installPackages
}
func (m mockAttempts) RemovePackages() int32 {
	return m.removePackages
}

func (m mockAttempts) ListServices() int32 {
	return m.listServices
}
func (m mockAttempts) IsEnabled() int32 {
	return m.isEnabled
}
func (m mockAttempts) DisableService() int32 {
	return m.disableService
}
func (m mockAttempts) EnableService() int32 {
	return m.enableService
}
func (m mockAttempts) StartService() int32 {
	return m.startService
}
func (m mockAttempts) StopSevice() int32 {
	return m.stopSevice
}
func (m mockAttempts) RestartService() int32 {
	return m.restartService
}

func (m mockAttempts) GetAll() int32 {
	return m.getAll
}
func (m mockAttempts) TSet() int32 {
	return m.tSet
}
func (m mockAttempts) Add() int32 {
	return m.add
}
func (m mockAttempts) Delete() int32 {
	return m.delete
}
func (m mockAttempts) CommitOrRevert() int32 {
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
func newMockAttempts(attempts int32) luci.Attempts {
	return mockAttempts{
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
		attempts,
	}
}
