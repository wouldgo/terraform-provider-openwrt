// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go tool mockgen -destination=../../../mocks/rpc.go -package=mocks -source=rpc.go -typed=true

// -build_constraint test
package luci

import (
	"context"
	"errors"
	"time"
)

var (
	ErrFloatExpected        = errors.New("value not a float64 type")
	ErrPackageNotFound      = errors.New("package not found")
	ErrPackagesNotSpecified = errors.New("no packages specified")

	ErrExecutionFailure = errors.New("execution returned value a failing result")

	ErrMissingRemoteBaseURL = errors.New("missing remote url")
	ErrMissingUsername      = errors.New("missing username")
	ErrMissingPassword      = errors.New("missing password")
)

type Timeouts interface {
	FsTimeouts
	OpkgTimeouts
	ServiceTimeouts
	SystemTimeouts

	Auth() time.Duration
}

type RPC interface {
	FsFacade
	OpkgFacade
	ServiceFacade
	SystemFacade
}

type RPCFactory interface {
	Get(ctx context.Context, url, username, password string, timeouts Timeouts) (RPC, error)
}
