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
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/testutil"
)

var _ apiCall = (*mockedCall)(nil)

type mockedCall struct {
	attempts    int32
	slotD       time.Duration
	timeoutD    time.Duration
	doFn        func(context.Context) (json.RawMessage, error)
	doCallCount int
}

func (m *mockedCall) attemptsAmount() int32        { return m.attempts }
func (m *mockedCall) timeoutAmount() time.Duration { return m.timeoutD }
func (m *mockedCall) slot() time.Duration          { return m.slotD }
func (m *mockedCall) do(ctx context.Context) (json.RawMessage, error) {
	m.doCallCount++
	return m.doFn(ctx)
}

type mockedTransformer[V any] struct {
	transformInErr     bool
	transformErr       error
	transformRes       V
	transformCallCount int
}

func (b *mockedTransformer[V]) Transform(json.RawMessage) (V, error) {
	b.transformCallCount++
	if b.transformInErr {
		var zero V
		return zero, b.transformErr
	}
	return b.transformRes, nil
}

func newMockedCall(attempts int32) *mockedCall {
	return &mockedCall{
		attempts: attempts,
		slotD:    0,
		timeoutD: time.Second,
	}
}

func TestAPICall_SuccessOnFirstAttempt(t *testing.T) {
	expected := "result"
	mock := newMockedCall(1)
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		return json.RawMessage(`"result"`), nil
	}
	transformer := &mockedTransformer[string]{
		transformRes: expected,
	}

	got, err := APICall(t.Context(), mock, transformer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
	if mock.doCallCount != 1 {
		t.Errorf("do called %d times, want 1", mock.doCallCount)
	}
}

func TestAPICall_SuccessAfterAttempts(t *testing.T) {
	callErr := errors.New("transient error")
	mock := newMockedCall(3)
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		if mock.doCallCount < 3 {
			return nil, callErr
		}
		return json.RawMessage(`"ok"`), nil
	}
	transformer := &mockedTransformer[string]{transformRes: "ok"}

	got, err := APICall(t.Context(), mock, transformer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("got %q, want %q", got, "ok")
	}
	if mock.doCallCount != 3 {
		t.Errorf("do called %d times, want 3", mock.doCallCount)
	}
}

func TestAPICall_TransformerErrorAttempts(t *testing.T) {
	transformErr := errors.New("bad payload")
	mock := newMockedCall(3)
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		return json.RawMessage(`"ok"`), nil
	}
	transformer := &mockedTransformer[string]{
		transformInErr: true,
		transformErr:   transformErr,
	}

	_, err := APICall(t.Context(), mock, transformer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMaxAttemptReached) {
		t.Errorf("expected ErrMaxAttemptReached in error chain, got: %v", err)
	}
	if mock.doCallCount != 3 {
		t.Errorf("do called %d times, want 3", mock.doCallCount)
	}
	if transformer.transformCallCount != 3 {
		t.Errorf("Transform called %d times, want 3", transformer.transformCallCount)
	}
}

func TestAPICall_ContextCancelledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	mock := &mockedCall{
		attempts: 5,
		slotD:    10 * time.Millisecond,
		timeoutD: time.Second,
	}
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		if mock.doCallCount == 1 {
			cancel()
		}
		return nil, errors.New("transient")
	}
	transformer := &mockedTransformer[string]{}

	_, err := APICall(ctx, mock, transformer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled in error chain, got: %v", err)
	}
}

func TestAPICall_MaxAttemptsExhausted(t *testing.T) {
	callErr := errors.New("always fails")
	mock := newMockedCall(3)
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		return nil, callErr
	}
	transformer := &mockedTransformer[string]{}

	_, err := APICall(t.Context(), mock, transformer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMaxAttemptReached) {
		t.Errorf("expected ErrMaxAttemptReached in error chain, got: %v", err)
	}
	if mock.doCallCount != 3 {
		t.Errorf("do called %d times, want 3", mock.doCallCount)
	}
}

func TestAPICall_ZeroAttempts(t *testing.T) {
	mock := newMockedCall(0)
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		return json.RawMessage(`"ok"`), nil
	}
	transformer := &mockedTransformer[string]{transformRes: "ok"}

	_, err := APICall(t.Context(), mock, transformer)
	if err == nil {
		t.Fatal("expected error with zero attempts, got nil")
	}
	if !errors.Is(err, ErrMaxAttemptReached) {
		t.Errorf("expected ErrMaxAttemptReached, got: %v", err)
	}
	if mock.doCallCount != 0 {
		t.Errorf("do called %d times, want 0", mock.doCallCount)
	}
}

func TestAPICall_ContextAlreadyCancelledOnEntry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	mock := newMockedCall(3)
	mock.doFn = func(_ context.Context) (json.RawMessage, error) {
		return nil, ctx.Err()
	}
	transformer := &mockedTransformer[string]{}

	_, err := APICall(ctx, mock, transformer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewBaseClientInError(t *testing.T) {
	httpClient := &http.Client{}
	t.Run("Wrong remote URL", func(t *testing.T) {
		_, err := NewBaseClient(
			httpClient,
			"username",
			"password",
			"127.0.0.1:3000",
			10*time.Second,
		)
		if err == nil || !errors.Is(err, ErrRemoteBaseURL) {
			t.Fatalf("expected NewBaseClient returning an error and the error must be a ErrRemoteBaseURL")
		}
	})

	t.Run("Nil http client remote URL", func(t *testing.T) {
		_, err := NewBaseClient(
			nil,
			"username",
			"password",
			"http://127.0.0.1:3000",
			10*time.Second,
		)
		if err == nil || !errors.Is(err, ErrNoHTTPClient) {
			t.Fatalf("expected NewBaseClient returning an error and the error must be a ErrNoHTTPClient")
		}
	})

}

func TestNewBaseClient(t *testing.T) {
	httpClient := &http.Client{}
	expectedURL := testutil.MustURL(t, "http://127.0.0.1:3000")
	baseClient, err := NewBaseClient(
		httpClient,
		"username",
		"password",
		"http://127.0.0.1:3000",
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("not expecting NewBaseClient returning an error")
	}
	if baseClient.username != "username" ||
		baseClient.password != "password" ||
		baseClient.authTimeout != 10*time.Second ||
		*baseClient.remoteBaseURL != *expectedURL ||
		baseClient.client != httpClient {
		t.Errorf("base client holds different values than the one passed to NewBaseClient")
	}
}

func TestPrepareCall(t *testing.T) {
	testCases := []struct {
		desc     string
		timeout  time.Duration
		attempts int32
		rpc      string
		method   string
		params   []any

		expectErr     bool
		expectedError error
	}{
		{
			desc:          "timeout equals to zero",
			timeout:       0,
			expectErr:     true,
			expectedError: ErrRpcTimeout,
		},
		{
			desc:      "negative attempts",
			timeout:   time.Second,
			attempts:  -1,
			rpc:       "test",
			method:    "m_test",
			expectErr: false,
		},
		{
			desc:          "no rpc",
			timeout:       time.Second,
			attempts:      0,
			rpc:           "",
			expectErr:     true,
			expectedError: ErrRpcCommand,
		},
		{
			desc:          "no method",
			timeout:       time.Second,
			attempts:      0,
			rpc:           "test",
			method:        "",
			expectErr:     true,
			expectedError: ErrRpcMethod,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			baseClient, _ := NewBaseClient(
				&http.Client{},
				"username",
				"password",
				"http://127.0.0.1:3000",
				10*time.Second,
			)

			ret, err := baseClient.PrepareCall(tC.timeout, tC.attempts, tC.rpc, tC.method, tC.params...)
			if err != nil && !tC.expectErr {
				t.Errorf("unexpected error raised: %s", err.Error())
			} else if !errors.Is(err, tC.expectedError) {
				t.Errorf(`expected error is different then returned error:
error: %s
expected: %s`, err.Error(), tC.expectedError.Error())
			}

			if err == nil && ret.attempts < 0 {
				t.Errorf("attempts must be greater or equals to zero")
			}
		})
	}
}

func TestAppendErr(t *testing.T) {
	testCases := []struct {
		desc        string
		errsCap     int
		errsForTest int
	}{
		{
			desc:        "append less errors than capacity",
			errsCap:     10,
			errsForTest: 5,
		},
		{
			desc:        "append same errors than capacity",
			errsCap:     10,
			errsForTest: 10,
		},
		{
			desc:        "append more errors than capacity",
			errsCap:     10,
			errsForTest: 15,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			errs := make([]error, 0, tC.errsCap)
			errsToAppend := make([]error, tC.errsForTest)
			for index := range tC.errsForTest {
				errsToAppend[index] = fmt.Errorf("err test %d", index)
			}
			for _, v := range errsToAppend {

				errs = appendErr(errs, v)
			}
			errsCap := cap(errs)
			if tC.errsCap != errsCap {
				t.Errorf("capacity changed: %v != %v", tC.errsCap, errsCap)
			}

			if tC.errsForTest > tC.errsCap {
				expectedErrs := errsToAppend[tC.errsForTest-tC.errsCap:]
				if len(errs) != len(expectedErrs) {
					t.Errorf("expected %d errors, got %d", len(expectedErrs), len(errs))
				}
				for i, err := range errs {
					if err.Error() != expectedErrs[i].Error() {
						t.Errorf("at index %d: expected %q, got %q", i, expectedErrs[i].Error(), err.Error())
					}
				}
			}
		})
	}
}

func TestCallAuth(t *testing.T) {
	expectedHost, authToken, expectedUsername, expectedPassword := testutil.TestingData(t)
	expectedRPC := "test"
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				err := testutil.SharedHTTPChecks(expectedHost, authToken, req)
				if err != nil {
					t.Fatalf("shared checks failed: %s", err.Error())
					return nil, err
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var rpcReq testutil.RPCRequest
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				switch req.URL.Path {
				case "/cgi-bin/luci/rpc/auth":
					response, err := testutil.HandleAuth(authToken, expectedUsername, expectedPassword, req, data)
					if err != nil {
						t.Errorf("handling auth in error: %s", err.Error())
						return nil, err
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(response)),
						Header:     make(http.Header),
					}, nil
				default:
					t.Fatalf("path \"%s\" not permitted in this test", req.URL.Path)
					return nil, errors.New("path not permitted")
				}
			},
		},
	}
	mockedCall := call{
		client:      mockedClient,
		username:    expectedUsername,
		password:    expectedPassword,
		attempts:    1,
		authTimeout: time.Second,
		currentURL:  expectedHost,
		rpc:         expectedRPC,
	}
	err := mockedCall.auth(t.Context())
	if err != nil {
		t.Errorf("auth call not expected to fail. Found %v", err)
	}
	if mockedCall.currentURL.Path != "cgi-bin/luci/rpc/"+expectedRPC {
		t.Errorf("expected returned auth url as cgi-bin/luci/rpc/%s. Found \"%s\"", expectedRPC, mockedCall.currentURL.Path)
	}
	if mockedCall.currentURL.Query().Get("auth") != authToken {
		t.Errorf("expected returned auth url with \"auth\" query string param to %s. Found \"%s\"", authToken, mockedCall.currentURL.Query().Get("auth"))
	}
	t.Logf("%v", mockedCall.currentURL)
}

func TestCallAuthTimeout(t *testing.T) {
	expectedHost, authToken, expectedUsername, expectedPassword := testutil.TestingData(t)
	expectedRPC := "test"
	expiredCtx, cancel := context.WithCancel(t.Context())
	cancel()
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				err := testutil.SharedHTTPChecks(expectedHost, authToken, req)
				if err != nil {
					t.Fatalf("shared checks failed: %s", err.Error())
					return nil, err
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var rpcReq testutil.RPCRequest
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				switch req.URL.Path {
				case "/cgi-bin/luci/rpc/auth":
					response, err := testutil.HandleAuth(authToken, expectedUsername, expectedPassword, req, data)
					if err != nil {
						t.Errorf("handling auth in error: %s", err.Error())
						return nil, err
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(response)),
						Header:     make(http.Header),
					}, nil
				default:
					t.Fatalf("path \"%s\" not permitted in this test", req.URL.Path)
					return nil, errors.New("path not permitted")
				}
			},
		},
	}
	mockedCall := call{
		client:      mockedClient,
		username:    expectedUsername,
		password:    expectedPassword,
		attempts:    1,
		authTimeout: time.Second,
		currentURL:  expectedHost,
		rpc:         expectedRPC,
	}
	err := mockedCall.auth(expiredCtx)
	if err == nil || !errors.Is(err, ErrAuthTimeout) {
		t.Errorf("auth call expected to fail with %v. Found %v", ErrAuthTimeout, err)
	}
}

func TestCallAuthHTTPError(t *testing.T) {
	expectedHost, authToken, expectedUsername, expectedPassword := testutil.TestingData(t)
	expectedRPC := "test"
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				err := testutil.SharedHTTPChecks(expectedHost, authToken, req)
				if err != nil {
					t.Fatalf("shared checks failed: %s", err.Error())
					return nil, err
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var rpcReq testutil.RPCRequest
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				switch req.URL.Path {
				case "/cgi-bin/luci/rpc/auth":
					_, err := testutil.HandleAuth(authToken, expectedUsername, expectedPassword, req, data)
					if err != nil {
						t.Errorf("handling auth in error: %s", err.Error())
						return nil, err
					}

					return nil, errors.ErrUnsupported
				default:
					t.Fatalf("path \"%s\" not permitted in this test", req.URL.Path)
					return nil, errors.New("path not permitted")
				}
			},
		},
	}
	mockedCall := call{
		client:      mockedClient,
		username:    expectedUsername,
		password:    expectedPassword,
		attempts:    1,
		authTimeout: time.Second,
		currentURL:  expectedHost,
		rpc:         expectedRPC,
	}
	err := mockedCall.auth(t.Context())
	if err == nil || !errors.Is(err, errors.ErrUnsupported) || !errors.Is(err, ErrHTTPRequestExecution) {
		t.Errorf("auth call expected to fail with %v and %v. Found %v", ErrAuthTimeout, ErrHTTPRequestExecution, err)
	}
}

func TestCallAuthHTTPFaultyResponse(t *testing.T) {
	expectedHost, authToken, expectedUsername, expectedPassword := testutil.TestingData(t)
	expectedRPC := "test"
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				err := testutil.SharedHTTPChecks(expectedHost, authToken, req)
				if err != nil {
					t.Fatalf("shared checks failed: %s", err.Error())
					return nil, err
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var rpcReq testutil.RPCRequest
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				switch req.URL.Path {
				case "/cgi-bin/luci/rpc/auth":
					_, err := testutil.HandleAuth(authToken, expectedUsername, expectedPassword, req, data)
					if err != nil {
						t.Errorf("handling auth in error: %s", err.Error())
						return nil, err
					}

					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewReader([]byte{})),
						Header:     make(http.Header),
					}, nil
				default:
					t.Fatalf("path \"%s\" not permitted in this test", req.URL.Path)
					return nil, errors.New("path not permitted")
				}
			},
		},
	}
	mockedCall := call{
		client:      mockedClient,
		username:    expectedUsername,
		password:    expectedPassword,
		attempts:    1,
		authTimeout: time.Second,
		currentURL:  expectedHost,
		rpc:         expectedRPC,
	}
	err := mockedCall.auth(t.Context())
	if err == nil || !errors.Is(err, ErrHTTPRequestExecution) {
		t.Errorf("auth call expected to fail with %v. Found %v", ErrHTTPRequestExecution, err)
	}
}

func TestActionHTTPMethod(t *testing.T) {
	expectedHTTPMethod := http.MethodPost
	expectedContentTypeHeader := "application/json; charset=utf-8"
	expectedURL := testutil.MustURL(t, "http://127.0.0.1:3000")
	expectedMethod := "test"
	expectedParams := []any{"param1", "param2"}
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Host != expectedURL.Host {
					t.Errorf("expected http url host is \"%s\". Found \"%s\"", expectedURL.Host, req.URL.Host)
				}
				if req.Method != expectedHTTPMethod {
					t.Errorf("expected http method is \"%s\". Found \"%s\"", expectedHTTPMethod, req.Method)
				}
				contentTypeHeaderValue := req.Header.Get("Content-Type")
				if contentTypeHeaderValue != expectedContentTypeHeader {
					t.Errorf("expected http header is \"%s\". Found \"%s\"", expectedContentTypeHeader, contentTypeHeaderValue)
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var (
					rpcReq  testutil.RPCRequest
					rpcResp testutil.RPCResponse
				)
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				if rpcReq.Method != expectedMethod {
					t.Errorf("expected method is \"%s\". Found \"%s\"", expectedMethod, rpcReq.Method)
				}

				stringyExpectedParams := make([]string, 0, len(expectedParams))
				stringyParams := make([]string, 0, len(rpcReq.Params))
				for _, aParam := range expectedParams {
					stringyExpectedParams = append(stringyExpectedParams, aParam.(string))
				}
				for _, aParam := range rpcReq.Params {
					stringyParams = append(stringyParams, aParam.(string))
				}

				if !slices.Equal(stringyExpectedParams, stringyParams) {
					t.Errorf("expected params are \"[%s]\". Found \"[%s]\"",
						strings.Join(stringyExpectedParams, ", "), strings.Join(stringyParams, ", "))
				}

				toReturn := json.RawMessage(`"TEST"`)
				rpcResp = testutil.RPCResponse{
					Result: &toReturn,
					Error:  nil,
				}
				responseBody, err := json.Marshal(&rpcResp)
				if err != nil {
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(responseBody)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	c := call{
		client:     mockedClient,
		currentURL: expectedURL,
		method:     expectedMethod,
		params:     expectedParams,
	}

	_, err := c.action(t.Context())
	if err != nil {
		t.Errorf("do exection expected to not fail. Found %v", err)
	}
	if c.currentURL != expectedURL {
		t.Errorf("do exection must not touch the expectedURL. Expected \"%s\". Found \"%s\"", expectedURL, c.currentURL)
	}
}

func TestActionErrCanceled(t *testing.T) {
	expectedURL := testutil.MustURL(t, "http://127.0.0.1:3000")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	c := call{
		method: "testing",
	}

	_, err := c.action(ctx)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("do execution must fail with context canceled error. Found %v", err)
	}
	if c.currentURL != nil {
		t.Errorf("do exection must return a nil expectedURL. Expected \"%s\". Found \"%s\"", expectedURL, c.currentURL)
	}
}

func TestActionErrHTTPRequestExecution(t *testing.T) {
	expectedURL := testutil.MustURL(t, "http://127.0.0.1:3000")
	expectedMethod := "test"
	expectedParams := []any{"param1", "param2"}
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Host != expectedURL.Host {
					t.Errorf("expected http url host is \"%s\". Found \"%s\"", expectedURL.Host, req.URL.Host)
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var (
					rpcReq  testutil.RPCRequest
					rpcResp testutil.RPCResponse
				)
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				if rpcReq.Method != expectedMethod {
					t.Errorf("expected method is \"%s\". Found \"%s\"", expectedMethod, rpcReq.Method)
				}

				stringyExpectedParams := make([]string, 0, len(expectedParams))
				stringyParams := make([]string, 0, len(rpcReq.Params))
				for _, aParam := range expectedParams {
					stringyExpectedParams = append(stringyExpectedParams, aParam.(string))
				}
				for _, aParam := range rpcReq.Params {
					stringyParams = append(stringyParams, aParam.(string))
				}

				if !slices.Equal(stringyExpectedParams, stringyParams) {
					t.Errorf("expected params are \"[%s]\". Found \"[%s]\"",
						strings.Join(stringyExpectedParams, ", "), strings.Join(stringyParams, ", "))
				}

				toReturn := json.RawMessage(`"TEST"`)
				rpcResp = testutil.RPCResponse{
					Result: &toReturn,
					Error:  nil,
				}
				responseBody, err := json.Marshal(&rpcResp)
				if err != nil {
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusConflict,
					Body:       io.NopCloser(bytes.NewReader(responseBody)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	c := call{
		client:     mockedClient,
		currentURL: expectedURL,
		method:     expectedMethod,
		params:     expectedParams,
	}

	_, err := c.action(t.Context())
	if err == nil || !errors.Is(err, ErrHTTPRequestExecution) {
		t.Errorf("do execution must fail with ErrHTTPRequestExecution error. Found %v", err)
	}
	if *expectedURL != *c.currentURL {
		t.Errorf("do exection changed the currentURL value. Expected \"%s\". Found \"%s\"", expectedURL, c.currentURL)
	}
}

func TestActionErrAuth(t *testing.T) {
	expectedURL := testutil.MustURL(t, "http://127.0.0.1:3000")
	expectedMethod := "test"
	expectedParams := []any{"param1", "param2"}
	mockedClient := &http.Client{
		Transport: &testutil.MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Host != expectedURL.Host {
					t.Errorf("expected http url host is \"%s\". Found \"%s\"", expectedURL.Host, req.URL.Host)
				}

				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				var (
					rpcReq  testutil.RPCRequest
					rpcResp testutil.RPCResponse
				)
				err = json.Unmarshal(data, &rpcReq)
				if err != nil {
					return nil, err
				}

				if rpcReq.Method != expectedMethod {
					t.Errorf("expected method is \"%s\". Found \"%s\"", expectedMethod, rpcReq.Method)
				}

				stringyExpectedParams := make([]string, 0, len(expectedParams))
				stringyParams := make([]string, 0, len(rpcReq.Params))
				for _, aParam := range expectedParams {
					stringyExpectedParams = append(stringyExpectedParams, aParam.(string))
				}
				for _, aParam := range rpcReq.Params {
					stringyParams = append(stringyParams, aParam.(string))
				}

				if !slices.Equal(stringyExpectedParams, stringyParams) {
					t.Errorf("expected params are \"[%s]\". Found \"[%s]\"",
						strings.Join(stringyExpectedParams, ", "), strings.Join(stringyParams, ", "))
				}

				toReturn := json.RawMessage(`"TEST"`)
				rpcResp = testutil.RPCResponse{
					Result: &toReturn,
					Error:  nil,
				}
				responseBody, err := json.Marshal(&rpcResp)
				if err != nil {
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(bytes.NewReader(responseBody)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	c := call{
		client:     mockedClient,
		currentURL: expectedURL,
		method:     expectedMethod,
		params:     expectedParams,
	}

	_, err := c.action(t.Context())
	if err == nil || !errors.Is(err, ErrAuth) {
		t.Errorf("do execution must fail with ErrAuth error. Found %v", err)
	}
	if *expectedURL != *c.currentURL {
		t.Errorf("do exection changed the currentURL value. Expected \"%s\". Found \"%s\"", expectedURL, c.currentURL)
	}
}
