// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
)

var (
	_ api.ApiDataTransformer[any]         = (*ignoreOutput)(nil)
	_ api.ApiDataTransformer[[]byte]      = base64Stringer{}
	_ api.ApiDataTransformer[any]         = opkgErrChecker{}
	_ api.ApiDataTransformer[PackageInfo] = opkgPackager{}
	_ api.ApiDataTransformer[bool]        = booleanTransformer{}
	_ api.ApiDataTransformer[[]string]    = stringSliceTransformer{}
	_ api.ApiDataTransformer[[]System]    = systemSliceTransformer{}
	_ api.ApiDataTransformer[string]      = stringTransformer{}
)

type ignoreOutput struct {
}

func (i ignoreOutput) Transform(json.RawMessage) (any, error) {
	var zero any
	return zero, nil
}

type base64Stringer struct {
}

func (base64Stringer) Transform(rawData json.RawMessage) ([]byte, error) {
	var s string
	if err := json.Unmarshal(rawData, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

type opkgErrChecker struct{}

func (opkgErrChecker) Transform(rawData json.RawMessage) (any, error) {
	var zero any

	var data []any
	if err := json.Unmarshal(rawData, &data); err != nil {
		return zero, errors.Join(ErrExecutionFailure, err)
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return zero, ErrFloatExpected
	}

	if retCasted != 0 {
		return zero, errors.Join(ErrExecutionFailure, fmt.Errorf("update packages returns %.0f", retCasted))
	}
	return zero, nil
}

type opkgPackager struct {
	pack string
}

func (o opkgPackager) Transform(rawData json.RawMessage) (PackageInfo, error) {
	var zero PackageInfo
	var data map[string]PackageInfo
	if err := json.Unmarshal(rawData, &data); err != nil {
		if err = json.Unmarshal(rawData, &[]bool{}); err != nil {

			return zero, errors.Join(ErrExecutionFailure, err)
		}

		return PackageInfo{
			Version: "",
			Status: Status{
				Installed: false,
			},
		}, nil
	}
	ret, ok := data[o.pack]
	if !ok {
		return zero, ErrPackageNotFound
	}
	return ret, nil
}

type booleanTransformer struct{}

func (booleanTransformer) Transform(rawData json.RawMessage) (bool, error) {
	var data bool
	if err := json.Unmarshal(rawData, &data); err != nil {
		return false, errors.Join(ErrExecutionFailure, err)
	}
	return data, nil
}

type stringSliceTransformer struct{}

func (stringSliceTransformer) Transform(rawData json.RawMessage) ([]string, error) {
	var data []string
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, errors.Join(ErrExecutionFailure, err)
	}
	return data, nil
}

type systemSliceTransformer struct {
	sections any
}

func (s systemSliceTransformer) Transform(rawData json.RawMessage) ([]System, error) {
	var data map[string]System
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data from the %v sections", s.sections)
	}

	return slices.Collect(maps.Values(data)), nil
}

type stringTransformer struct{}

func (stringTransformer) Transform(rawData json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(rawData, &s); err != nil {
		return "", err
	}
	return s, nil
}
