// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http_test

import (
	"context"
	"errors"
	"testing"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci_http"
)

func TestGetErrs(t *testing.T) {
	factory, err := luci_http.NewHTTPRPCFactory(luci_http.HTTPRPCConfiguration{})
	if err != nil {
		t.Fatal("unexpected factory creation in error")
	}

	alreadyCancelledCtx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = factory.Get(alreadyCancelledCtx, "", "", "", nil)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("expecting %v error, got %v", context.Canceled, err)
	}

	_, err = factory.Get(t.Context(), "", "", "", nil)
	if err == nil || !errors.Is(err, luci.ErrMissingRemoteBaseURL) {
		t.Errorf("expecting %v error, got %v", luci.ErrMissingRemoteBaseURL, err)
	}

	_, err = factory.Get(t.Context(), "http://openwrt.lan", "", "", nil)
	if err == nil || !errors.Is(err, luci.ErrMissingUsername) {
		t.Errorf("expecting %v error, got %v", luci.ErrMissingUsername, err)
	}

	_, err = factory.Get(t.Context(), "http://openwrt.lan", "root", "", nil)
	if err == nil || !errors.Is(err, luci.ErrMissingPassword) {
		t.Errorf("expecting %v error, got %v", luci.ErrMissingPassword, err)
	}

	_, err = factory.Get(t.Context(), "http://openwrt.lan", "root", "password", nil)
	if err == nil || !errors.Is(err, luci.ErrMissingTimeouts) {
		t.Errorf("expecting %v error, got %v", luci.ErrMissingTimeouts, err)
	}
}
