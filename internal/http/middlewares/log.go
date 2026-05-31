// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http_middlewares

import (
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func WithLogging() func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			resp, err := next.RoundTrip(req)
			if err == nil {
				tflog.Info(req.Context(), "http request", map[string]any{
					"method":   req.Method,
					"url":      req.URL.String(),
					"status":   resp.StatusCode,
					"duration": time.Since(start),
				})
			} else {
				tflog.Error(req.Context(), "http request in error", map[string]any{
					"method":   req.Method,
					"url":      req.URL.String(),
					"duration": time.Since(start),
				})
			}
			return resp, err
		})
	}
}
