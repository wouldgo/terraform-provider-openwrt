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
	http_transformers "github.com/foxboron/terraform-provider-openwrt/internal/http/transformers"
	http_transport "github.com/foxboron/terraform-provider-openwrt/internal/http/transport"
)

// fakeResponse builds a minimal *http.Response the caller can return from RoundTrip.
func fakeResponse(t *testing.T, statusCode int, body string) *http.Response {
	defer func() {
		t.Logf("status code %d: %s", statusCode, body)
	}()
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
//	Action 1   → 200     retryable, error from the response body given by the http_transformers logic
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
		Result: json.RawMessage(`[0]`),
	})

	opkgLockErr, _ := json.Marshal(jsonRPCResponseBody{
		Result: json.RawMessage(`[65280,\"\",\"Collected errors:\\u000a * opkg_conf_load: Could not lock /var/lock/opkg.lock: Resource temporarily unavailable.\\u000a\"]`),
	})

	maxSecondsWaitInRoundTripper := big.NewInt(5)
	testingRoundTripper := http_middlewares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		randSecs, err := rand.Int(rand.Reader, maxSecondsWaitInRoundTripper)
		if err != nil {
			return nil, err
		}
		waitAmount := time.Duration(randSecs.Int64()) * time.Second

		authToken := req.URL.Query().Get("auth")

		if strings.HasSuffix(req.URL.Path, "/rpc/auth") {
			if authToken != "" {
				t.Error("auth request must not have the auth query string value set")
			}

			n := authCallCount.Add(1)
			t.Logf("auth call #%d will wait %s", n, waitAmount)
			time.Sleep(waitAmount)
			switch n {
			case 1:
				return fakeResponse(t, http.StatusOK, `{"result":"token-v1"}`), nil
			case 2:
				return fakeResponse(t, http.StatusOK, `{"result":"token-v2"}`), nil
			default:
				t.Errorf("unexpected auth call #%d", n)
				return fakeResponse(t, http.StatusInternalServerError, ""), nil
			}
		}

		if authToken == "" {
			t.Error("request must have the auth query string value set")
		}

		if !strings.HasSuffix(req.URL.Path, "/rpc/RPC_test") {
			t.Error("url must have the suffix /rpc/RPC_test")
		}

		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("error on reading request body: %v", err)
		}

		unmarshalledRequestBody := jsonRPCRequestBody{}
		err = json.Unmarshal(reqBody, &unmarshalledRequestBody)
		if err != nil {
			t.Errorf("error unmarshalling request body to a jsonRPCRequestBody: %v", err)
		}

		if unmarshalledRequestBody.Method != "METHOD_test" {
			t.Error("request body method must be METHOD_test")
		}

		if unmarshalledRequestBody.Params != nil {
			t.Error("request body params must be nil")
		}

		n := actionCallCount.Add(1)
		t.Logf("action call #%d will wait %s", n, waitAmount)
		time.Sleep(waitAmount)
		switch {
		case n == 1: // attempt 0 → 502
			return fakeResponse(t, http.StatusBadGateway, ""), nil
		case n == 2: // attempt 1 → 200 but with a result with specific error (managed by the transformer)
			return fakeResponse(t, http.StatusOK, string(opkgLockErr)), nil
		case n == 3: // attempt 2 → 403; AuthTransport re-auths and retries (hits n==4)
			return fakeResponse(t, http.StatusForbidden, ""), nil
		case n == 4: // attempt 2 internal retry after 403 → 502
			return fakeResponse(t, http.StatusBadGateway, ""), nil
		case n >= 5 && n <= 9: // attempts 3-7 → 502
			return fakeResponse(t, http.StatusBadGateway, ""), nil
		case n == 10: // attempt 8 → success
			return fakeResponse(t, http.StatusOK, string(successBody)), nil
		default:
			t.Errorf("unexpected action call #%d", n)
			return fakeResponse(t, http.StatusInternalServerError, ""), nil
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

	_, err = Call(
		t.Context(),
		baseClient,
		timeout,
		"RPC_test",
		"METHOD_test",
		http_transformers.OpkgErrCheckerTransformer)
	if err != nil {
		t.Fatalf("Call() returned an error: %v", err)
	}

	if n := authCallCount.Load(); n != 2 {
		t.Errorf("auth calls = %d, want 2", n)
	}

	if n := actionCallCount.Load(); n != 10 {
		t.Errorf("action calls = %d, want 10", n)
	}
}
