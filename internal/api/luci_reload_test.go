// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/testutil"
)

// TestAPICall_LuciReloadScenario replays the real-world openwrt/luci failure
// mode that motivated the retry + exponential backoff work:
//
//   Attempt 0:
//     auth  → 200 OK (token issued, <1s)
//     action → hangs for ciActionHang, then returns 502 Bad Gateway
//
//   Attempts 1–7  (7 attempts):
//     auth  → 502 Bad Gateway  (luci daemon is reloading)
//
//   Attempts 8–17  (10 attempts):
//     auth  → 200 OK
//     action → 500 Internal Server Error  (generic error, e.g. opkg still busy)
//
//   Attempt 18+:
//     auth  → 200 OK
//     action → 200 OK + valid JSON result  ← success
//
// Constants below are scaled for fast CI.  Swap in the "real hardware" values
// (commented out) to profile actual backoff behaviour against a device.

const (
	ciTimeout     = 200 * time.Millisecond
	ciAuthTimeout = 40 * time.Millisecond
	ciActionHang  = 20 * time.Millisecond

	phaseActionHangAttempt = 0
	phase502AuthStart      = 1
	phase502AuthEnd        = 7
	phaseGenericErrStart   = 8
	phaseGenericErrEnd     = 17
	phaseSuccessStart      = 18
)

// ── HTTP response helpers ──────────────────────────────────────────────────

func authOKResponse(token string) *http.Response {
	body := fmt.Sprintf(`{"id":1,"result":%q,"error":null}`, token)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func rpcOKResponse(result string) *http.Response {
	body := fmt.Sprintf(`{"id":1,"result":%s,"error":null}`, result)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func errResponse(code int) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

// ── Test ───────────────────────────────────────────────────────────────────

func TestAPICall_LuciReloadScenario(t *testing.T) {
	// attemptIdx is incremented by countingCall *after* each do() returns,
	// so inside RoundTripFunc it always reflects the current in-flight attempt.
	var attemptIdx atomic.Int32

	transport := &testutil.MockRoundTripper{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			isAuth := strings.Contains(req.URL.Path, "rpc/auth")
			attempt := int(attemptIdx.Load())

			t.Logf("[roundtrip] attempt=%d isAuth=%v path=%s", attempt, isAuth, req.URL.Path)

			switch {

			// ── Phase 0: first attempt ─────────────────────────────────────
			case attempt == phaseActionHangAttempt && isAuth:
				return authOKResponse("token-phase0"), nil

			case attempt == phaseActionHangAttempt && !isAuth:
				time.Sleep(ciActionHang)
				return errResponse(http.StatusBadGateway), nil

			// ── Phase 1: luci reloading — auth endpoint is down ───────────
			case attempt >= phase502AuthStart && attempt <= phase502AuthEnd && isAuth:
				return errResponse(http.StatusBadGateway), nil

			// ── Phase 2: luci back, opkg still busy ───────────────────────
			case attempt >= phaseGenericErrStart && attempt <= phaseGenericErrEnd && isAuth:
				return authOKResponse(fmt.Sprintf("token-phase2-%d", attempt)), nil

			case attempt >= phaseGenericErrStart && attempt <= phaseGenericErrEnd && !isAuth:
				return errResponse(http.StatusInternalServerError), nil

			// ── Phase 3: everything works ──────────────────────────────────
			case attempt >= phaseSuccessStart && isAuth:
				return authOKResponse("token-phase3"), nil

			case attempt >= phaseSuccessStart && !isAuth:
				return rpcOKResponse(`"ok"`), nil

			default:
				return nil, fmt.Errorf("unexpected request: attempt=%d isAuth=%v path=%s",
					attempt, isAuth, req.URL.Path)
			}
		},
	}

	baseURL, err := url.Parse("http://openwrt.test:8080")
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}

	c := &call{
		client:      &http.Client{Transport: transport},
		username:    "root",
		password:    "test",
		authTimeout: ciAuthTimeout,
		currentURL:  baseURL,
		timeout:     ciTimeout,
		rpc:         "sys",
		method:      "hostname",
		params:      []any{},
	}

	// countingCall wraps the real *call to keep attemptIdx in sync.
	wrapped := &countingCall{inner: c, counter: &attemptIdx}

	transformer := &mockedTransformer[string]{transformRes: "hostname-result"}

	// Outer deadline: generous enough that a misconfiguration produces a clean
	// timeout failure rather than a hang.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	result, callErr := APICall(ctx, wrapped, transformer)
	elapsed := time.Since(start)

	if callErr != nil {
		t.Fatalf("APICall returned unexpected error after %s: %v", elapsed, callErr)
	}
	if result != "hostname-result" {
		t.Errorf("unexpected result: got %q want %q", result, "hostname-result")
	}

	finalAttempt := int(attemptIdx.Load()) - 1 // 0-indexed
	t.Logf("succeeded on attempt %d, elapsed=%s, transformer calls=%d",
		finalAttempt, elapsed, transformer.transformCallCount)

	if finalAttempt < phaseSuccessStart {
		t.Errorf("expected first success on attempt >= %d, got it on attempt %d",
			phaseSuccessStart, finalAttempt)
	}
	if transformer.transformCallCount == 0 {
		t.Error("transformer.Transform was never called")
	}
	t.Logf("%d", attemptIdx.Load())
}

// countingCall wraps a *call, delegating everything and incrementing the
// atomic counter after each do() so RoundTripFunc reads the right attempt index.
type countingCall struct {
	inner   *call
	counter *atomic.Int32
}

func (cc *countingCall) timeoutAmount() time.Duration { return cc.inner.timeoutAmount() }
func (cc *countingCall) slot() time.Duration          { return cc.inner.slot() }
func (cc *countingCall) do(ctx context.Context) (json.RawMessage, error) {
	result, err := cc.inner.do(ctx)
	cc.counter.Add(1)
	return result, err
}
