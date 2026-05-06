// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

var (
	_ error = (*jsonRPCResponseError)(nil)

	ErrNoHTTPClient  = errors.New("missing http client")
	ErrRemoteBaseURL = errors.New("remote base URL is not valid")
	ErrRpcTimeout    = errors.New("missing rpc timeout")
	ErrRpcCommand    = errors.New("missing rpc command")
	ErrRpcMethod     = errors.New("missing rpc method")

	ErrHTTPBodyRead = errors.New("http request in error")

	ErrAuth                 = errors.New("authencation in error")
	ErrAuthTimeout          = errors.New("authentication timeout")
	ErrRpcExecution         = errors.New("rpc execution error")
	ErrMaxAttemptReached    = errors.New("max attempts reached")
	ErrMarshal              = errors.New("json marshal in error")
	ErrParsing              = errors.New("parsing error")
	ErrHttpRequestCreation  = errors.New("http request creation in error")
	ErrHTTPRequestExecution = errors.New("http request execution in error")
	ErrUnMarshal            = errors.New("json unmarshal in error")
	ErrEmptyResult          = errors.New("empty reply as result")
)

type BaseClient struct {
	username, password string
	authTimeout        time.Duration
	remoteBaseURL      *url.URL
	client             *http.Client
}

func NewBaseClient(
	client *http.Client,
	username, password, remoteBaseURL string,
	authTimeout time.Duration,
) (*BaseClient, error) {
	if client == nil {
		return nil, ErrNoHTTPClient
	}
	remoteURL, err := url.Parse(remoteBaseURL)
	if err != nil {
		return nil, errors.Join(err, ErrRemoteBaseURL)
	}

	return &BaseClient{
		client:        client,
		username:      username,
		password:      password,
		remoteBaseURL: remoteURL,
		authTimeout:   authTimeout,
	}, nil
}

func (b *BaseClient) PrepareCall(
	timeout time.Duration,
	rpc, method string,
	params ...any,
) (*call, error) {
	if timeout == 0 {
		return nil, ErrRpcTimeout
	}
	if rpc == "" {
		return nil, ErrRpcCommand
	}
	if method == "" {
		return nil, ErrRpcMethod
	}

	return &call{
		client:      b.client,
		username:    b.username,
		password:    b.password,
		authTimeout: b.authTimeout,
		currentURL:  b.remoteBaseURL,
		timeout:     timeout,
		rpc:         rpc,
		method:      method,
		params:      params,
	}, nil
}
