// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

//go:generate go tool mockgen -destination=../../mocks/api.go -package=mocks -source=api.go -typed=true

// -build_constraint test
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ClientFactory interface {
	Get(url string) (Client, error)
}

type Client interface {
	SystemFacade
	FsFacade
	OpkgFacade
	InitFacade

	Auth(ctx context.Context, username, password string) error
}

var (
	_ ClientFactory = (*clientFactory)(nil)
	_ Client        = (*client)(nil)

	ErrMissingUrl           = fmt.Errorf("missing remote url")
	ErrParsing              = fmt.Errorf("parsing error")
	ErrMarshal              = fmt.Errorf("json marshal in error")
	ErrHttpRequestCreation  = fmt.Errorf("http request creation in error")
	ErrHttpRequestExecution = fmt.Errorf("http request execution in error")
	ErrUnMarshal            = fmt.Errorf("json unmarshal in error")
	ErrAuth                 = fmt.Errorf("authencation in error")
	ErrEmptyResult          = fmt.Errorf("empty reply as result")
	ErrRpcCommand           = fmt.Errorf("missing rpc command")
	ErrRpcMethod            = fmt.Errorf("missing rpc method")
	ErrRpcExecution         = fmt.Errorf("rpc execution error")

	ErrFloatExpected    = fmt.Errorf("value not a float64 type")
	ErrExecutionFailure = fmt.Errorf("execution returned value a failing result")
	ErrPackageNotFound  = fmt.Errorf("package not found")

	ErrPackagesNotSpecified = fmt.Errorf("no packages specified")
)

type clientFactory struct {
}

func NewClientFactory() (ClientFactory, error) {
	return &clientFactory{}, nil
}

func (cf *clientFactory) Get(url string) (Client, error) {
	return newClient(url)
}

type client struct {
	*uci
	*fs
	*opkg
	*service
	url    string
	client *http.Client
}

func newClient(url string) (Client, error) {
	if url == "" {
		return nil, ErrMissingUrl
	}

	client := &client{
		url:    url,
		client: &http.Client{},
	}

	return client, nil
}

func (c *client) Auth(ctx context.Context, username, password string) error {
	innerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	b, err := json.Marshal(&struct {
		Id     int      `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Id:     1,
		Method: "login",
		Params: []string{username, password},
	})
	if err != nil {
		return errors.Join(ErrMarshal, err)
	}
	u, err := url.JoinPath(c.url, "cgi-bin/luci/rpc/auth")
	if err != nil {
		return errors.Join(ErrParsing, err)
	}

	req, err := http.NewRequestWithContext(innerCtx, http.MethodPost, u, bytes.NewBuffer(b))
	if err != nil {
		return errors.Join(ErrHttpRequestCreation, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Join(ErrHttpRequestExecution, err)
	}

	var data struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return errors.Join(ErrUnMarshal, fmt.Errorf("failed to read authentication response body: %w", err))
	}

	if data.Error != "" {
		return errors.Join(ErrAuth, fmt.Errorf("authentication error: %s", data.Error))
	}

	if data.Result == "" {
		return ErrEmptyResult
	}

	c.uci = &uci{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}
	c.fs = &fs{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}
	c.opkg = &opkg{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}
	c.service = &service{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}

	return nil
}

type jsonRPCRequestBody struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

var _ error = (*jsonRPCResponseError)(nil)

type jsonRPCResponseError struct {
	Code    float64 `json:"code"`
	Message string  `json:"message"`
}

func (j *jsonRPCResponseError) Error() string {
	return fmt.Sprintf("rpc call in error: %.0f: %s", j.Code, j.Message)
}

type jsonRPCResponseBody struct {
	Error  *jsonRPCResponseError `json:"error"`
	Result *json.RawMessage      `json:"result"`
}

func call(
	ctx context.Context, client *http.Client,
	remoteUrl, token, rpc, method string, params []any,
) (json.RawMessage, error) {
	innerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if remoteUrl == "" {
		return nil, ErrMissingUrl
	}
	if token == "" {
		return nil, errors.Join(ErrAuth, fmt.Errorf("no auth is performed against %s", remoteUrl))
	}
	if rpc == "" {
		return nil, ErrRpcCommand
	}
	if method == "" {
		return nil, ErrRpcMethod
	}
	u, err := url.Parse(remoteUrl)
	if err != nil {
		return nil, errors.Join(ErrParsing, err)
	}
	u.Path = fmt.Sprintf("cgi-bin/luci/rpc/%s", rpc)
	q := u.Query()
	q.Add("auth", token)
	u.RawQuery = q.Encode()

	requestBody, err := json.Marshal(&jsonRPCRequestBody{
		Method: method,
		Params: params,
	})
	if err != nil {
		return nil, errors.Join(ErrMarshal, err)
	}
	req, err := http.NewRequestWithContext(innerCtx, http.MethodPost, u.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Join(ErrHttpRequestCreation, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	tflog.Debug(ctx, "start - request call to remote", map[string]interface{}{
		"host":    req.URL.Host,
		"path":    req.URL.Path,
		"method":  req.Method,
		"request": string(requestBody),
	})
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Join(ErrHttpRequestExecution, err)
	}

	tflog.Debug(ctx, "end - request call to remote", map[string]interface{}{
		"host":    req.URL.Host,
		"path":    req.URL.Path,
		"method":  req.Method,
		"request": string(requestBody),
		"response": map[string]interface{}{
			"statusCode": resp.StatusCode,
		},
	})

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(ErrHttpRequestExecution, fmt.Errorf("request %+v replied with %d", req, resp.StatusCode))
	}
	var responseBody jsonRPCResponseBody
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close() //nolint:errcheck
	if err = decoder.Decode(&responseBody); err != nil {
		return nil, errors.Join(ErrUnMarshal, err)
	}
	if responseBody.Error != nil {
		return nil, errors.Join(ErrRpcExecution, responseBody.Error)
	}

	if responseBody.Result == nil {
		return nil, ErrEmptyResult
	}

	return *responseBody.Result, nil
}

// Purge the sections from the anonymous things
func purgeFields(d any) (any, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(b, &objmap); err != nil {
		return nil, err
	}
	delete(objmap, ".name")
	delete(objmap, ".anonymous")
	delete(objmap, ".type")
	return objmap, nil
}
