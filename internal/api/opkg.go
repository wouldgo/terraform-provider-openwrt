// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultUpdatePackagesTimeout  time.Duration = 30 * time.Second
	defaultCheckPackageTimeout                  = 30 * time.Second
	defaultInstallPackagesTimeout               = 30 * time.Second
	defaultRemovePackagesTimeout                = 30 * time.Second
)

type OpkgTimeouts interface {
	UpdatePackages() time.Duration
	CheckPackage() time.Duration
	InstallPackages() time.Duration
	RemovePackages() time.Duration
}

type OpkgTimeoutsModel struct {
	UpdatePackagesTimeout  types.String `tfsdk:"update_packages"`
	CheckPackageTimeout    types.String `tfsdk:"check_package"`
	InstallPackagesTimeout types.String `tfsdk:"install_packages"`
	RemovePackagesTimeout  types.String `tfsdk:"remove_packages"`
}

type OpkgFacade interface {
	UpdatePackages(ctx context.Context) error
	CheckPackage(ctx context.Context, pack string) (*PackageInfo, error)
	InstallPackages(ctx context.Context, packages ...string) error
	RemovePackages(ctx context.Context, packages ...string) error
}

type opkgTimeouts struct {
	updatePackagesTimeout,
	checkPackageTimeout,
	installPackagesTimeout,
	removePackagesTimeout time.Duration
}

func (opkgT *opkgTimeouts) UpdatePackages() time.Duration {
	return opkgT.updatePackagesTimeout
}

func (opkgT *opkgTimeouts) CheckPackage() time.Duration {
	return opkgT.checkPackageTimeout
}

func (opkgT *opkgTimeouts) InstallPackages() time.Duration {
	return opkgT.installPackagesTimeout
}

func (opkgT *opkgTimeouts) RemovePackages() time.Duration {
	return opkgT.removePackagesTimeout
}

var (
	_ OpkgFacade   = (*opkg)(nil)
	_ WithSession  = (*opkg)(nil)
	_ OpkgTimeouts = (*opkgTimeouts)(nil)

	opkgTimeoutSchemaAttribute = schema.SingleNestedAttribute{
		MarkdownDescription: `Opkg operations timeout configuration`,
		Description:         `Opkg operations timeout configuration`,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"update_packages": schema.StringAttribute{
				MarkdownDescription: `Update packages RPC timeout value`,
				Description:         `Update packages RPC timeout value`,
				Optional:            true,
			},
			"check_package": schema.StringAttribute{
				MarkdownDescription: `Check package RPC timeout value`,
				Description:         `Check package RPC timeout value`,
				Optional:            true,
			},
			"install_packages": schema.StringAttribute{
				MarkdownDescription: `Install packages RPC timeout value`,
				Description:         `Install packages RPC timeout value`,
				Optional:            true,
			},
			"remove_packages": schema.StringAttribute{
				MarkdownDescription: `Remove packages RPC timeout value`,
				Description:         `Remove packages RPC timeout value`,
				Optional:            true,
			},
		},
	}
)

func parseOpkgTimeouts(ctx context.Context, t *TimeoutsModel) (OpkgTimeouts, error) {
	updatePackagesTimeout := defaultUpdatePackagesTimeout
	checkPackageTimeout := defaultCheckPackageTimeout
	installPackagesTimeout := defaultInstallPackagesTimeout
	removePackagesTimeout := defaultRemovePackagesTimeout

	if t != nil && t.Opkg != nil && !t.Opkg.UpdatePackagesTimeout.IsNull() {
		parsedUpdatePackagesTimeout, err := time.ParseDuration(t.Opkg.UpdatePackagesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		updatePackagesTimeout = parsedUpdatePackagesTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: update_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default update_packages config")
	}

	if t != nil && t.Opkg != nil && !t.Opkg.CheckPackageTimeout.IsNull() {
		parsedCheckPackageTimeout, err := time.ParseDuration(t.Opkg.CheckPackageTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		checkPackageTimeout = parsedCheckPackageTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: check_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default check_packages config")
	}

	if t != nil && t.Opkg != nil && !t.Opkg.InstallPackagesTimeout.IsNull() {
		parsedInstallPackagesTimeout, err := time.ParseDuration(t.Opkg.InstallPackagesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		installPackagesTimeout = parsedInstallPackagesTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: install_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default install_packages config")
	}

	if t != nil && t.Opkg != nil && !t.Opkg.RemovePackagesTimeout.IsNull() {
		parsedRemovePackagesTimeout, err := time.ParseDuration(t.Opkg.RemovePackagesTimeout.ValueString())
		if err != nil {
			return nil, err
		}

		removePackagesTimeout = parsedRemovePackagesTimeout
		tflog.Debug(ctx, "opkg - parse timeout configuration: remove_packages config parsed")
	} else {
		tflog.Debug(ctx, "opkg - parse timeout configuration: default remove_packages config")
	}

	return &opkgTimeouts{
		updatePackagesTimeout,
		checkPackageTimeout,
		installPackagesTimeout,
		removePackagesTimeout,
	}, nil
}

type opkg struct {
	timeouts OpkgTimeouts

	token  string
	url    *string
	client *http.Client
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

func (c *opkg) SetToken(ctx context.Context, token string) error {
	c.token = token
	return nil
}

func (c *opkg) UpdatePackages(ctx context.Context) error {
	result, err := call(ctx, c.client, c.timeouts.UpdatePackages(),
		*c.url, c.token,
		"ipkg", "update", []any{})
	if err != nil {
		return err
	}

	var data []any
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return ErrFloatExpected
	}

	if retCasted != 0 {
		return errors.Join(ErrExecutionFailure, fmt.Errorf("update packages returns %.0f", retCasted))
	}
	return nil
}

func (c *opkg) CheckPackage(ctx context.Context, pack string) (*PackageInfo, error) {
	result, err := call(ctx, c.client, c.timeouts.CheckPackage(),
		*c.url, c.token,
		"ipkg", "status", []any{pack})
	if err != nil {
		return nil, err
	}

	var data map[string]PackageInfo
	if err = json.Unmarshal(result, &data); err != nil {
		if err = json.Unmarshal(result, &[]bool{}); err != nil {

			return nil, errors.Join(ErrUnMarshal, err)
		}

		return &PackageInfo{
			Version: "",
			Status: Status{
				Installed: false,
			},
		}, nil
	}
	ret, ok := data[pack]
	if !ok {
		return nil, ErrPackageNotFound
	}
	return &ret, nil
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
	result, err := call(ctx, c.client, c.timeouts.InstallPackages(),
		*c.url, c.token, "ipkg", "install", toApi)
	if err != nil {
		return err
	}

	var data []any
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return ErrFloatExpected
	}

	if retCasted != 0 {
		return errors.Join(ErrExecutionFailure, fmt.Errorf("install packages returns %.0f", retCasted))
	}
	return nil
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
	result, err := call(ctx, c.client, c.timeouts.RemovePackages(),
		*c.url, c.token, "ipkg", "remove", toApi)
	if err != nil {
		return err
	}

	var data []any
	if err = json.Unmarshal(result, &data); err != nil {
		return errors.Join(ErrUnMarshal, err)
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return ErrFloatExpected
	}

	if retCasted != 0 {
		return errors.Join(ErrExecutionFailure, fmt.Errorf("remove packages returns %.0f", retCasted))
	}
	return nil
}
