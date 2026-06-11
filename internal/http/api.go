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
	"maps"
	"net/http"
	"net/url"
	"time"

	http_middlewares "github.com/foxboron/terraform-provider-openwrt/internal/http/middlewares"
	http_transformers "github.com/foxboron/terraform-provider-openwrt/internal/http/transformers"
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
			FadeEnd:    0.6,
		}
	} else {
		baseClient.backoffDurationStrategy = backoffDurationStrategy
	}
	return baseClient, nil
}

func Call[T any](
	ctx context.Context,
	b *BaseClient,
	timeout time.Duration,
	rpc, method string,
	trans http_transformers.ApiDataTransformer[T],
	params ...any,
) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, errors.Join(ErrRpcTimeout, err)
	}

	req, err := prepareRequest(
		ctx,
		b.remoteBaseURL,
		rpc,
		method,
		params,
	)
	if err != nil {
		return zero, err
	}

	resp, err := retry(
		ctx,
		b,
		timeout,
		trans,
		req,
		rpc,
		method,
	)
	if err != nil {
		return zero, errors.Join(ErrRpcExecution, err)
	}
	return resp, nil
}

func extractBodyContent[T any](
	ctx context.Context,
	trans http_transformers.ApiDataTransformer[T],
	resp *http.Response,
) (T, error) {
	var zero T
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
			return zero, errors.Join(ErrHTTPBodyRead, fmt.Errorf("faulty error response read"))
		}

		errResp := fmt.Errorf("request %s - %s replied with %d: %s", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode, string(errorResponse))

		return zero, errors.Join(ErrHTTPRequestExecution, errResp)
	}

	var responseBody jsonRPCResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return zero, errors.Join(ErrUnMarshal, err)
	}
	if responseBody.Error != NoJSONRPCError {
		return zero, errors.Join(ErrRpcExecution, responseBody.Error)
	}

	if responseBody.Result == nil {
		return zero, ErrEmptyResult
	}

	return trans.Transform(responseBody.Result)
}

func retry[T any](
	ctx context.Context,
	b *BaseClient,
	timeout time.Duration,
	trans http_transformers.ApiDataTransformer[T],
	req *http.Request,
	rpc, method string,
) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, errors.Join(err, ErrRpcTimeout)
	}

	innerCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		attempt      int32
		resp         *http.Response
		totalIO            = 100 * time.Millisecond
		attemptsDone int32 = 0
		start              = time.Now()
		ioErr        error
	)

	for attempt = 0; ; attempt++ {
		clonedReq, err := http_middlewares.CloneRequest(req)
		if err != nil {
			return zero, err
		}

		ioStart := time.Now()
		resp, ioErr = b.Do(clonedReq)
		ioElapsed := time.Since(ioStart)
		if ioErr != nil {
			return zero, errors.Join(err, ErrTransport)
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
			data, ioErr := extractBodyContent(innerCtx, trans, resp)
			if ioErr == nil {
				tflog.Info(innerCtx, "call completed", loggingData)
				return data, nil
			} else {
				loggingData = map[string]any{
					"attempt":      attempt,
					"rpc":          rpc,
					"method":       method,
					"statusCode":   resp.StatusCode,
					"responseTime": ioElapsed.String(),

					"error": ioErr.Error(),
				}
				tflog.Error(innerCtx, "extracting body content gives error", loggingData)
			}
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

		backoffLoggingData := map[string]any{
			"elapsed":    elapsed.String(),
			"remaining":  (timeout - elapsed).String(),
			"average IO": avgIO.String(),
			"backoff":    backoffDuration.String(),
		}
		maps.Copy(backoffLoggingData, loggingData)
		tflog.Debug(innerCtx, "backoff", backoffLoggingData)

		select {
		case <-innerCtx.Done():
			return zero, errors.Join(ErrRetryDeadlineExceeded, fmt.Errorf("stopped after %d attempts: %w", attempt, ioErr))
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
