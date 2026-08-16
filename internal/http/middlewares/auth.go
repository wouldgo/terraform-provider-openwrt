// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http_middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_                  http.RoundTripper = (*AuthTransport)(nil)
	ErrAuth                              = errors.New("authencation in error")
	ErrAuthEmptyResult                   = errors.New("authentication result is empty")

	ErrAuthTimeout = errors.New("authentication timeout")
	ErrMarshal     = errors.New("auth json marshal in error")
	ErrUnMarshal   = errors.New("auth json unmarshal in error")

	ErrAuthHttpRequestCreation  = errors.New("http auth request creation in error")
	ErrAuthHTTPRequestExecution = errors.New("http request execution in error")
	ErrAuthHTTPBodyRead         = errors.New("http request in error")
)

func WithAuth(
	username, password string,
	authTimeout time.Duration,
) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &AuthTransport{
			base:     next,
			username: username,
			password: password,
			timeout:  authTimeout,
		}
	}
}

type AuthTransport struct {
	base               http.RoundTripper
	username, password string
	timeout            time.Duration

	token string
}

func (a *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := a.base
	if base == nil {
		base = http.DefaultTransport
	}

	err := a.handleAuth(req.Context(), req.URL)
	if err != nil {
		return nil, errors.Join(ErrAuth, err)
	}

	q := req.URL.Query()
	q.Set("auth", a.token)
	req.URL.RawQuery = q.Encode()

	clonedReq, err := CloneRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := base.RoundTrip(req)
	if err == nil && resp.StatusCode == http.StatusForbidden {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		q := clonedReq.URL.Query()
		q.Del("auth")
		clonedReq.URL.RawQuery = q.Encode()
		a.token = ""

		return a.RoundTrip(clonedReq)
	}
	return resp, err
}

func (a *AuthTransport) handleAuth(
	ctx context.Context,
	authURL *url.URL,
) error {
	base := a.base
	if base == nil {
		base = http.DefaultTransport
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(ErrAuthTimeout, err)
	}

	if a.token != "" {
		return nil
	}

	innerCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	b, err := json.Marshal(&struct {
		Id     int      `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Id:     1,
		Method: "login",
		Params: []string{a.username, a.password},
	})
	if err != nil {
		return errors.Join(ErrMarshal, err)
	}

	theAuthURL := CloneURL(authURL)
	theAuthURL.Path = "cgi-bin/luci/rpc/auth"
	// #nosec G107 G704
	authReq, err := http.NewRequestWithContext(
		innerCtx,
		http.MethodPost,
		theAuthURL.String(),
		bytes.NewBuffer(b),
	)
	if err != nil {
		return errors.Join(ErrAuthHttpRequestCreation, err)
	}
	authReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	authResp, err := base.RoundTrip(authReq)
	if err != nil {
		return errors.Join(ErrAuthHTTPRequestExecution, err)
	}

	defer func(authResp *http.Response) {
		err := authResp.Body.Close()
		if err != nil {
			tflog.Error(ctx, "closing of the response body in error", map[string]any{
				"error": err.Error(),
			})
		}
	}(authResp)

	if authResp.StatusCode != http.StatusOK {
		errorResponse, err := io.ReadAll(authResp.Body)
		if err != nil {
			return errors.Join(ErrAuthHTTPBodyRead, fmt.Errorf("faulty error response read"))
		}

		return errors.Join(
			ErrAuthHTTPRequestExecution,
			fmt.Errorf("request %s - %s replied with %d: %s",
				authReq.Method,
				authURL.String(),
				authResp.StatusCode,
				string(errorResponse),
			),
		)
	}

	var data struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(authResp.Body).Decode(&data); err != nil {
		return errors.Join(ErrUnMarshal, fmt.Errorf("failed to read authentication response body: %w", err))
	}

	if data.Error != "" {
		return errors.Join(ErrAuth, fmt.Errorf("authentication error: %s", data.Error))
	}

	if data.Result == "" {
		return ErrAuthEmptyResult
	}

	tflog.Debug(ctx, "authentication performed", map[string]any{
		"host":     authURL.Host,
		"username": a.username,
	})

	a.token = data.Result
	return nil
}
