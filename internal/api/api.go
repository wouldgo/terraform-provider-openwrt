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

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ClientFactory interface {
	ParseTimeouts(ctx context.Context, timeouts *TimeoutsModel) (Timeouts, error)
	Get(ctx context.Context, url string, timeouts Timeouts) (Client, error)
}

type Timeouts interface {
	FsTimeouts
	OpkgTimeouts
	ServiceTimeouts
	SystemTimeouts

	Auth() time.Duration
}

type Client interface {
	FsFacade
	OpkgFacade
	ServiceFacade
	SystemFacade

	Auth(ctx context.Context, username, password string) error
}

type WithSession interface {
	SetToken(ctx context.Context, token string) error
}

type TimeoutsModel struct {
	Auth types.String `tfsdk:"auth"`

	Fs      *FsTimeoutsModel      `tfsdk:"fs"`
	Opkg    *OpkgTimeoutsModel    `tfsdk:"opkg"`
	Service *ServiceTimeoutsModel `tfsdk:"service"`
	System  *SystemTimeoutsModel  `tfsdk:"uci"`
}

type timeouts struct {
	FsTimeouts
	OpkgTimeouts
	ServiceTimeouts
	SystemTimeouts

	authTimeout time.Duration
}

func (t *timeouts) Auth() time.Duration {
	return t.authTimeout
}

const defaultAuthTimeout = 5 * time.Second

var (
	_ ClientFactory = (*clientFactory)(nil)
	_ Client        = (*client)(nil)
	_ Timeouts      = (*timeouts)(nil)

	TimeoutSchemaAttribute = schema.SingleNestedAttribute{
		MarkdownDescription: "Timeout configuration for the specific RPC calls. The main purpose of this optional configuration is to fine tune the default timeouts for longer API interaction (e.g. update packages, list packages, ...)",
		Description:         "Timeout configuration for the specific RPC calls",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"auth": schema.StringAttribute{
				MarkdownDescription: `Authentication RPC timeout value`,
				Description:         `Authentication RPC timeout value`,
				Optional:            true,
			},
			"fs":      fsTimeoutSchemaAttribute,
			"opkg":    opkgTimeoutSchemaAttribute,
			"service": serviceTimeoutSchemaAttribute,
			"uci":     uciTimeoutSchemaAttribute,
		},
	}

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

func (cf *clientFactory) ParseTimeouts(ctx context.Context, t *TimeoutsModel) (Timeouts, error) {
	authTimeout := defaultAuthTimeout
	if t != nil && !t.Auth.IsNull() {
		parsedAuthTimeout, err := time.ParseDuration(t.Auth.ValueString())
		if err != nil {
			return nil, err
		}

		authTimeout = parsedAuthTimeout
		tflog.Debug(ctx, "parse timeout configuration: auth config parsed")
	} else {
		tflog.Debug(ctx, "parse timeout configuration: default auth config")
	}

	fs, err := parseFsTimeouts(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("error parsing fs timeouts: %w", err)
	}

	opkg, err := parseOpkgTimeouts(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("error parsing opkg timeouts: %w", err)
	}

	service, err := parseServiceTimeouts(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("error parsing service timeouts: %w", err)
	}

	system, err := parseSystemTimeouts(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("error parsing system timeouts: %w", err)
	}

	return &timeouts{
		fs,
		opkg,
		service,
		system,
		authTimeout,
	}, nil
}

func (cf *clientFactory) Get(ctx context.Context, url string, t Timeouts) (Client, error) {
	return newClient(url, t)
}

type client struct {
	FsFacade
	OpkgFacade
	ServiceFacade
	SystemFacade

	needToken []WithSession

	url      *string
	client   *http.Client
	timeouts Timeouts
}

func newClient(url string, t Timeouts) (Client, error) {
	if url == "" {
		return nil, ErrMissingUrl
	}
	httpClient := &http.Client{}
	remoteUrl := &url

	fs := &fs{
		timeouts: t,
		url:      remoteUrl,
		client:   httpClient,
	}

	opkg := &opkg{
		timeouts: t,
		url:      remoteUrl,
		client:   httpClient,
	}

	service := &service{
		timeouts: t,
		url:      remoteUrl,
		client:   httpClient,
	}

	system := &system{
		timeouts: t,
		url:      remoteUrl,
		client:   httpClient,
	}

	client := &client{
		FsFacade:      fs,
		OpkgFacade:    opkg,
		ServiceFacade: service,
		SystemFacade:  system,
		needToken: []WithSession{
			fs, opkg, service, system,
		},
		timeouts: t,
		url:      remoteUrl,
		client:   httpClient,
	}

	return client, nil
}

func (c *client) Auth(ctx context.Context, username, password string) error {
	tflog.Debug(ctx, "authentication", map[string]interface{}{
		"url":      c.url,
		"username": username,
	})
	innerCtx, cancel := context.WithTimeout(ctx, c.timeouts.Auth())
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
	u, err := url.JoinPath(*c.url, "cgi-bin/luci/rpc/auth")
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

	if resp.StatusCode != http.StatusOK {
		return errors.Join(ErrHttpRequestExecution, fmt.Errorf("authentication request %+v replied with %d", req, resp.StatusCode))
	}

	defer resp.Body.Close() //nolint:errcheck

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

	tflog.Debug(ctx, "authentication performed", map[string]interface{}{
		"url":      c.url,
		"usernema": username,
	})

	token := data.Result
	for _, v := range c.needToken {
		err = v.SetToken(ctx, token)
		if err != nil {
			return fmt.Errorf("error on setting token for fs: %w", err)
		}
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
	ctx context.Context, client *http.Client, timeout time.Duration,
	remoteUrl, token, rpc, method string, params []any,
) (json.RawMessage, error) {
	innerCtx, cancel := context.WithTimeout(ctx, timeout)
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
	defer resp.Body.Close() //nolint:errcheck

	var responseBody jsonRPCResponseBody
	if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
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
