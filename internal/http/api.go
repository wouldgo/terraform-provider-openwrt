// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http

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

	http_middlewares "github.com/foxboron/terraform-provider-openwrt/internal/http/middlewares"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	NoJSONRPCError error = jsonRPCResponseError{}

	ErrRemoteBaseURL = errors.New("remote base URL is not valid")
	ErrRpcTimeout    = errors.New("missing rpc timeout")
	ErrRpcCommand    = errors.New("missing rpc command")
	ErrRpcMethod     = errors.New("missing rpc method")

	ErrHTTPBodyRead = errors.New("http request in error")

	ErrRpcExecution         = errors.New("rpc execution error")
	ErrMarshal              = errors.New("json marshal in error")
	ErrUnMarshal            = errors.New("json unmarshal in error")
	ErrParsing              = errors.New("parsing error")
	ErrHttpRequestCreation  = errors.New("http request creation in error")
	ErrHTTPRequestExecution = errors.New("http request execution in error")

	ErrTransport             = errors.New("network error")
	ErrRetryDeadlineExceeded = errors.New("retry deadline exceeded")

	ErrEmptyResult = errors.New("empty reply as result")
)

type jsonRPCResponseBody struct {
	Error  jsonRPCResponseError `json:"error"`
	Result json.RawMessage      `json:"result"`
}

type jsonRPCRequestBody struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type jsonRPCResponseError struct {
	Code    float64 `json:"code"`
	Message string  `json:"message"`
}

func (j jsonRPCResponseError) Error() string {
	return fmt.Sprintf("rpc call in error: %.0f: %s", j.Code, j.Message)
}

type BaseClient struct {
	*http.Client
	backoffDurationStrategy BackoffDurationStrategy
	remoteBaseURL           *url.URL
}

func NewBaseClient(
	httpClient *http.Client,
	backoffDurationStrategy BackoffDurationStrategy,
	remoteBaseURL string,
) (*BaseClient, error) {
	remoteURL, err := url.Parse(remoteBaseURL)
	if err != nil {
		return nil, errors.Join(err, ErrRemoteBaseURL)
	}

	baseClient := &BaseClient{
		Client:        httpClient,
		remoteBaseURL: remoteURL,
	}

	if backoffDurationStrategy == nil {
		baseClient.backoffDurationStrategy = MinMaxRetriesBackOffStrategy{
			MinBackOff: 100 * time.Millisecond,
			MinRetries: 5,
			MaxRetries: 13,
		}
	}
	return baseClient, nil
}

func (b *BaseClient) Call(
	ctx context.Context,
	timeout time.Duration,
	rpc, method string,
	params ...any,
) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, ErrRpcTimeout)
	}

	req, err := prepareRequest(
		ctx,
		b.remoteBaseURL,
		rpc,
		method,
		params,
	)
	if err != nil {
		return nil, err
	}

	resp, err := b.retry(ctx, timeout, req, rpc, method)
	if err != nil {
		return nil, errors.Join(ErrRpcExecution, err)
	}

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			tflog.Error(ctx, "closing of the response body in error", map[string]any{
				"error": err.Error(),
			})
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		errorResponse, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Join(ErrHTTPBodyRead, fmt.Errorf("faulty error response read"))
		}

		errResp := fmt.Errorf("request %s - %s replied with %d: %s", req.Method, b.remoteBaseURL.String(), resp.StatusCode, string(errorResponse))

		return nil, errors.Join(ErrHTTPRequestExecution, errResp)
	}

	var responseBody jsonRPCResponseBody
	if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, errors.Join(ErrUnMarshal, err)
	}
	if responseBody.Error != NoJSONRPCError {
		return nil, errors.Join(ErrRpcExecution, responseBody.Error)
	}

	if responseBody.Result == nil {
		return nil, ErrEmptyResult
	}

	return responseBody.Result, nil
}

func (b *BaseClient) retry(
	ctx context.Context,
	timeout time.Duration,
	req *http.Request,
	rpc, method string,
) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, ErrRpcTimeout)
	}

	innerCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		attempt      int32
		resp         *http.Response
		totalIO            = 100 * time.Millisecond
		attemptsDone int32 = 0
		start              = time.Now()
	)

	for attempt = 0; ; attempt++ {
		clonedReq, err := http_middlewares.CloneRequest(req)
		if err != nil {
			return nil, err
		}

		ioStart := time.Now()
		resp, err = b.Do(clonedReq)
		ioElapsed := time.Since(ioStart)
		if err != nil {
			return nil, errors.Join(err, ErrTransport)
		}

		attemptsDone++
		totalIO += ioElapsed
		avgIO := totalIO / time.Duration(attemptsDone+1)

		loggingData := map[string]any{
			"attempt":      attempt,
			"rpc":          rpc,
			"method":       method,
			"statusCode":   resp.StatusCode,
			"responseTime": ioElapsed.String(),
		}
		if !isRetryable(resp.StatusCode) {
			tflog.Info(innerCtx, "call completed", loggingData)
			return resp, nil
		}

		tflog.Info(innerCtx, "attempt failed", loggingData)
		clonedReqURL := clonedReq.URL.Query()
		if authToken := clonedReqURL.Get("auth"); authToken != "" {
			q := req.URL.Query()
			q.Set("auth", authToken)
			req.URL.RawQuery = q.Encode()
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		elapsed := time.Since(start)
		backoffDuration := b.backoffDurationStrategy.NextBackoffDuration(innerCtx, elapsed, timeout, avgIO)
		tflog.Debug(innerCtx, "backoff", map[string]any{
			"attempt":    attempt,
			"elapsed":    elapsed.String(),
			"remaining":  (timeout - elapsed).String(),
			"average IO": avgIO.String(),
			"backoff":    backoffDuration.String(),
		})

		select {
		case <-innerCtx.Done():
			return nil, fmt.Errorf("stopped after %d attempts: %w", attempt, ErrRetryDeadlineExceeded)
		case <-time.After(backoffDuration):
		}
	}
}

func prepareRequest(
	ctx context.Context,
	baseUrl *url.URL,
	rpc, method string,
	params []any,
) (*http.Request, error) {
	newUrl := http_middlewares.CloneURL(baseUrl)
	newUrl.Path = fmt.Sprintf("cgi-bin/luci/rpc/%s", rpc)
	requestBody, err := json.Marshal(&jsonRPCRequestBody{
		Method: method,
		Params: params,
	})
	if err != nil {
		return nil, errors.Join(ErrMarshal, err)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		newUrl.String(),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, errors.Join(ErrHttpRequestCreation, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return req, nil
}

func isRetryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}
