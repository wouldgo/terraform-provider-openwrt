// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	http_middlewares "github.com/foxboron/terraform-provider-openwrt/internal/http/middlewares"
	http_transport "github.com/foxboron/terraform-provider-openwrt/internal/http/transport"
)

// fakeResponse builds a minimal *http.Response the caller can return from RoundTrip.
func fakeResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// TestRetryScenario verifies the following call sequence:
//
//	Auth   0   → 200 OK  (token "token-v1" issued)
//	Action 0   → 502     retryable, backoff
//	Action 1   → 502     retryable, backoff
//	Action 2   → 403     AuthTransport clears token, retries internally:
//	Auth   1   → 200 OK  (token "token-v2" issued)
//	Action 2b  → 502     retryable (call() sees this as attempt 2's result)
//	Action 3-7 → 502     retryable, backoff
//	Action 8   → 200 OK  valid JSON-RPC result  ← success
func TestRetryScenario(t *testing.T) {
	var (
		authCallCount   atomic.Int32
		actionCallCount atomic.Int32
	)

	successBody, _ := json.Marshal(jsonRPCResponseBody{
		Result: json.RawMessage(`"ok"`),
	})

	maxSecondsWaitInRoundTripper := big.NewInt(5)
	testingRoundTripper := http_middlewares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		randSecs, err := rand.Int(rand.Reader, maxSecondsWaitInRoundTripper)
		if err != nil {
			return nil, err
		}
		waitAmount := time.Duration(randSecs.Int64()) * time.Second
		t.Logf("%s will wait %s", req.URL.Path, waitAmount)
		time.Sleep(waitAmount)

		authToken := req.URL.Query().Get("auth")

		if strings.HasSuffix(req.URL.Path, "/rpc/auth") {
			if authToken != "" {
				t.Error("auth request must not have the auth query string value set")
			}

			n := authCallCount.Add(1)
			switch n {
			case 1:
				return fakeResponse(http.StatusOK,
					`{"result":"token-v1"}`), nil
			case 2:
				return fakeResponse(http.StatusOK,
					`{"result":"token-v2"}`), nil
			default:
				t.Errorf("unexpected auth call #%d", n)
				return fakeResponse(http.StatusInternalServerError, ""), nil
			}
		}

		if authToken == "" {
			t.Error("request must have the auth query string value set")
		}
		n := actionCallCount.Add(1)
		switch {
		case n == 1: // attempt 0 → 502
			return fakeResponse(http.StatusBadGateway, ""), nil
		case n == 2: // attempt 1 → 502
			return fakeResponse(http.StatusBadGateway, ""), nil
		case n == 3: // attempt 2 → 403; AuthTransport re-auths and retries (hits n==4)
			return fakeResponse(http.StatusForbidden, ""), nil
		case n == 4: // attempt 2 internal retry after 403 → 502
			return fakeResponse(http.StatusBadGateway, ""), nil
		case n >= 5 && n <= 9: // attempts 3-7 → 502
			return fakeResponse(http.StatusBadGateway, ""), nil
		case n == 10: // attempt 8 → success
			return fakeResponse(http.StatusOK, string(successBody)), nil
		default:
			t.Errorf("unexpected action call #%d", n)
			return fakeResponse(http.StatusInternalServerError, ""), nil
		}
	})

	dt, err := http_transport.NewDynamicTransport(testingRoundTripper)
	if err != nil {
		t.Fatalf("NewDynamicTransport: %v", err)
	}

	dt.Add("authentication", http_middlewares.WithAuth(
		"admin", "password",
		500*time.Millisecond,
	))

	baseClient, err := NewBaseClient(
		&http.Client{Transport: dt},
		nil,
		"http://openwrt.local",
	)
	if err != nil {
		t.Fatalf("newBaseClient: %v", err)
	}

	timeout := 1 * time.Minute

	t.Logf("wait for %s", timeout)

	result, err := baseClient.Call(t.Context(), timeout, "sys", "hostname")
	if err != nil {
		t.Fatalf("call() returned unexpected error: %v", err)
	}

	var got string
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if got != "ok" {
		t.Errorf("result = %q, want %q", got, "ok")
	}

	if n := authCallCount.Load(); n != 2 {
		t.Errorf("auth calls = %d, want 2", n)
	}

	if n := actionCallCount.Load(); n != 10 {
		t.Errorf("action calls = %d, want 10", n)
	}
}
