// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ apiCall = (*call)(nil)
)

type apiCall interface {
	timeoutAmount() time.Duration
	slot() time.Duration
	do(context.Context) (json.RawMessage, error)
}

type ApiDataTransformer[V any] interface {
	Transform(json.RawMessage) (V, error)
}

type call struct {
	client             *http.Client
	username, password string
	authTimeout        time.Duration

	currentURL  *url.URL
	timeout     time.Duration
	rpc, method string
	params      []any
}

func (c *call) timeoutAmount() time.Duration {
	return c.timeout
}

func (c *call) slot() time.Duration {
	fivePercent := float64(c.timeout) * 0.05
	return time.Duration(math.Round(fivePercent))
}

func (c *call) do(ctx context.Context) (json.RawMessage, error) {
	var zero json.RawMessage
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	err := c.auth(ctx)
	if err != nil {
		return zero, err
	}

	data, err := c.action(ctx)
	if err != nil {
		return zero, err
	}
	return data, nil
}

func (c *call) auth(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return errors.Join(err, ErrAuthTimeout)
	}

	authToken := c.currentURL.Query().Get("auth")
	if authToken != "" {
		return nil
	}

	tflog.Debug(ctx, "authentication", map[string]interface{}{
		"host":     c.currentURL.Host,
		"username": c.username,
	})
	innerCtx, cancel := context.WithTimeout(ctx, c.authTimeout)
	defer cancel()
	b, err := json.Marshal(&struct {
		Id     int      `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Id:     1,
		Method: "login",
		Params: []string{c.username, c.password},
	})
	if err != nil {
		return errors.Join(ErrMarshal, err)
	}

	c.currentURL.Path = fmt.Sprintf("cgi-bin/luci/rpc/%s", "auth")
	req, err := http.NewRequestWithContext(
		innerCtx,
		http.MethodPost,
		c.currentURL.String(),
		bytes.NewBuffer(b),
	)
	if err != nil {
		return errors.Join(ErrHttpRequestCreation, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Join(ErrHTTPRequestExecution, err)
	}

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			tflog.Error(ctx, "closing of the response body in error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		errorResponse, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Join(ErrHTTPBodyRead, fmt.Errorf("faulty error response read"))
		}

		return errors.Join(ErrHTTPRequestExecution, fmt.Errorf("request %s - %s replied with %d: %s", req.Method, c.currentURL.String(), resp.StatusCode, string(errorResponse)))
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

	tflog.Debug(ctx, "authentication performed", map[string]interface{}{
		"host":     c.currentURL.Host,
		"username": c.username,
	})

	q := c.currentURL.Query()
	q.Add("auth", data.Result)
	c.currentURL.RawQuery = q.Encode()

	return nil
}

func (c *call) action(ctx context.Context) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, ErrRpcTimeout)
	}

	c.currentURL.Path = fmt.Sprintf("cgi-bin/luci/rpc/%s", c.rpc)
	requestBody, err := json.Marshal(&jsonRPCRequestBody{
		Method: c.method,
		Params: c.params,
	})
	if err != nil {
		return nil, errors.Join(ErrMarshal, err)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.currentURL.String(),
		bytes.NewBuffer(requestBody),
	)
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
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Join(ErrHTTPRequestExecution, err)
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

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			tflog.Error(ctx, "closing of the response body in error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		errorResponse, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Join(ErrHTTPBodyRead, fmt.Errorf("faulty error response read"))
		}

		errResp := fmt.Errorf("request %s - %s replied with %d: %s", req.Method, c.currentURL.String(), resp.StatusCode, string(errorResponse))
		if resp.StatusCode == http.StatusForbidden {
			q := c.currentURL.Query()
			q.Del("auth")
			c.currentURL.RawQuery = q.Encode()
			return nil, errors.Join(ErrAuth, errResp)
		}

		return nil, errors.Join(ErrHTTPRequestExecution, errResp)
	}

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

type jsonRPCResponseBody struct {
	Error  *jsonRPCResponseError `json:"error"`
	Result *json.RawMessage      `json:"result"`
}

type jsonRPCRequestBody struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type jsonRPCResponseError struct {
	Code    float64 `json:"code"`
	Message string  `json:"message"`
}

func (j *jsonRPCResponseError) Error() string {
	return fmt.Sprintf("rpc call in error: %.0f: %s", j.Code, j.Message)
}

func appendErr(errs []error, err ...error) []error {
	for _, e := range err {
		if len(errs) < cap(errs) {
			errs = append(errs, e)
		} else {
			copy(errs, errs[1:])
			errs[len(errs)-1] = e
		}
	}
	return errs
}

func APICall[V any](
	ctx context.Context,
	call apiCall,
	transformer ApiDataTransformer[V],
) (V, error) {
	var (
		zero    V
		attempt int32
	)
	errs := make([]error, 0, 10)

	for attempt = 0; ; attempt++ {
		rawMessage, err := call.do(ctx)
		if err == nil {
			toReturn, err := transformer.Transform(rawMessage)
			if err == nil {
				return toReturn, nil
			}
			errs = appendErr(errs, err)
		} else {
			errs = appendErr(errs, err)
		}

		exp := attempt + 1
		slotDuration := call.slot()
		var maxWindow uint32
		if slotDuration > 0 {
			maxWindow = max(uint32(call.timeoutAmount()/slotDuration), 1)
		} else {
			maxWindow = 1
		}
		candidateWindowSize := uint32(1) << exp
		if candidateWindowSize == 0 {
			candidateWindowSize = 1
		}
		windowSize := min(candidateWindowSize, maxWindow)
		k := rand.Int64N(int64(windowSize))
		backoffDuration := time.Duration(k) * call.slot()
		select {
		case <-ctx.Done():
			errs = appendErr(errs, ErrMaxAttemptReached)
			return zero, fmt.Errorf("%w after %d attempts: %w", ctx.Err(), attempt, errors.Join(errs...))
		case <-time.After(backoffDuration):
		}
	}
}
