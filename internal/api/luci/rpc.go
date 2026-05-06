// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go tool mockgen -destination=../../../mocks/rpc.go -package=mocks -source=rpc.go -typed=true

// -build_constraint test
package luci

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
)

var (
	_ RPCFactory = rpcFactory{}
	_ RPC        = (*rpc)(nil)

	ErrMissingRemoteBaseURL = errors.New("missing remote url")
	ErrMissingUsername      = errors.New("missing username")
	ErrMissingPassword      = errors.New("missing password")

	ErrExecutionFailure = errors.New("execution returned value a failing result")

	ErrFloatExpected   = errors.New("value not a float64 type")
	ErrPackageNotFound = errors.New("package not found")

	ErrPackagesNotSpecified = errors.New("no packages specified")
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

type rpcFactory struct {
	httpClient *http.Client
}

func NewHTTPRPCFactory(httpClient *http.Client) (RPCFactory, error) {
	return rpcFactory{
		httpClient: httpClient,
	}, nil
}

func (cf rpcFactory) Get(ctx context.Context, remoteBaseURL, username, password string, timeouts Timeouts) (RPC, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if remoteBaseURL == "" {
		return nil, ErrMissingRemoteBaseURL
	}
	if username == "" {
		return nil, ErrMissingUsername
	}
	if password == "" {
		return nil, ErrMissingPassword
	}

	baseClient, err := api.NewBaseClient(
		cf.httpClient,
		username,
		password,
		remoteBaseURL,
		timeouts.Auth(),
	)
	if err != nil {
		return nil, err
	}

	fs := &fs{
		baseClient,
		timeouts,
	}

	opkg := &opkg{
		baseClient,
		timeouts,
	}

	service := &service{
		baseClient,
		timeouts,
	}

	system := &system{
		baseClient,
		timeouts,
	}

	return &rpc{
		fs:      fs,
		opkg:    opkg,
		service: service,
		system:  system,
	}, nil
}

type rpc struct {
	*fs
	*opkg
	*service
	*system
}
