// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	this_http "github.com/foxboron/terraform-provider-openwrt/internal/http"
	http_middlewares "github.com/foxboron/terraform-provider-openwrt/internal/http/middlewares"
	http_transport "github.com/foxboron/terraform-provider-openwrt/internal/http/transport"
)

var (
	_ luci.RPCFactory = (*httpRPCFactory)(nil)
	_ luci.RPC        = (*httpRPC)(nil)
)

type HTTPRPCConfiguration struct {
	RoundTripper            http.RoundTripper
	BackoffDurationStrategy this_http.BackoffDurationStrategy
}

type httpRPCFactory struct {
	dynamicTransport        *http_transport.DynamicTransport
	backoffDurationStrategy this_http.BackoffDurationStrategy
	httpClient              *http.Client
}

type httpRPC struct {
	*fs
	*opkg
	*service
	*system
}

func NewHTTPRPCFactory(
	httpRPCConfiguration HTTPRPCConfiguration,
) (luci.RPCFactory, error) {
	dt, err := http_transport.NewDynamicTransport(httpRPCConfiguration.RoundTripper)
	if err != nil {
		return nil, fmt.Errorf("error on creating transport: %w", err)
	}

	httpClient := &http.Client{
		Transport: dt,
	}
	return &httpRPCFactory{
		dynamicTransport:        dt,
		backoffDurationStrategy: httpRPCConfiguration.BackoffDurationStrategy,
		httpClient:              httpClient,
	}, nil
}

func (cf *httpRPCFactory) Get(ctx context.Context,
	remoteBaseURL, username, password string,
	timeouts luci.Timeouts,
) (luci.RPC, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if remoteBaseURL == "" {
		return nil, luci.ErrMissingRemoteBaseURL
	}
	if username == "" {
		return nil, luci.ErrMissingUsername
	}
	if password == "" {
		return nil, luci.ErrMissingPassword
	}
	if timeouts == nil {
		return nil, luci.ErrMissingTimeouts
	}

	cf.dynamicTransport.Add("logging", http_middlewares.WithLogging())
	cf.dynamicTransport.Add("authentication", http_middlewares.WithAuth(
		username, password,
		timeouts.Auth(),
	))

	baseClient, err := this_http.NewBaseClient(
		cf.httpClient,
		cf.backoffDurationStrategy,
		remoteBaseURL,
	)
	if err != nil {
		return nil, err
	}

	return &httpRPC{
		fs: &fs{
			baseClient,
			timeouts,
		},
		opkg: &opkg{
			baseClient,
			timeouts,
		},
		service: &service{
			baseClient,
			timeouts,
		},
		system: &system{
			baseClient,
			timeouts,
		},
	}, nil
}
