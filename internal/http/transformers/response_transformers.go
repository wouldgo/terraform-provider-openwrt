// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http_transformers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
)

var (
	IgnoreOutputTransformer   ApiDataTransformer[any]              = IgnoreOutput{}
	Base64StringerTransformer ApiDataTransformer[[]byte]           = Base64Stringer{}
	OpkgErrCheckerTransformer ApiDataTransformer[any]              = OpkgErrChecker{}
	_                         ApiDataTransformer[luci.PackageInfo] = OpkgPackager{}
	BooleanTransformer        ApiDataTransformer[bool]             = Boolean{}
	StringSliceTransformer    ApiDataTransformer[[]string]         = StringSlice{}
	_                         ApiDataTransformer[[]luci.System]    = SystemSlice{}
	StringTransformer         ApiDataTransformer[string]           = String{}
)

type ApiDataTransformer[V any] interface {
	Transform(json.RawMessage) (V, error)
}

type IgnoreOutput struct {
}

func (i IgnoreOutput) Transform(json.RawMessage) (any, error) {
	var zero any
	return zero, nil
}

type Base64Stringer struct {
}

func (Base64Stringer) Transform(rawData json.RawMessage) ([]byte, error) {
	var s string
	if err := json.Unmarshal(rawData, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

type OpkgErrChecker struct{}

func (OpkgErrChecker) Transform(rawData json.RawMessage) (any, error) {
	var zero any

	var data []any
	if err := json.Unmarshal(rawData, &data); err != nil {
		return zero, err
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return zero, luci.ErrFloatExpected
	}

	if retCasted != 0 {
		return zero, fmt.Errorf("update packages returns %.0f", retCasted)
	}
	return zero, nil
}

type OpkgPackager struct {
	Package string
}

func (o OpkgPackager) Transform(rawData json.RawMessage) (luci.PackageInfo, error) {
	var zero luci.PackageInfo
	var data map[string]luci.PackageInfo
	if err := json.Unmarshal(rawData, &data); err != nil {
		if err = json.Unmarshal(rawData, &[]bool{}); err != nil {

			return zero, err
		}

		return luci.PackageInfo{
			Version: "",
			Status: luci.Status{
				Installed: false,
			},
		}, nil
	}
	ret, ok := data[o.Package]
	if !ok {
		return zero, luci.ErrPackageNotFound
	}
	return ret, nil
}

type Boolean struct{}

func (Boolean) Transform(rawData json.RawMessage) (bool, error) {
	var data bool
	if err := json.Unmarshal(rawData, &data); err != nil {
		return false, err
	}
	return data, nil
}

type StringSlice struct{}

func (StringSlice) Transform(rawData json.RawMessage) ([]string, error) {
	var data []string
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}
	return data, nil
}

type SystemSlice struct {
	Sections any
}

func (s SystemSlice) Transform(rawData json.RawMessage) ([]luci.System, error) {
	var data map[string]luci.System
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data from the %v sections", s.Sections)
	}

	return slices.Collect(maps.Values(data)), nil
}

type String struct{}

func (String) Transform(rawData json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(rawData, &s); err != nil {
		return "", err
	}
	return s, nil
}
