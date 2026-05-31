// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http_middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func CloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}

	c := *u
	return &c
}

func CloneRequest(req *http.Request) (*http.Request, error) {
	cloned := req.Clone(req.Context())

	if req.Body == nil || req.Body == http.NoBody {
		return cloned, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	cloned.Body = io.NopCloser(bytes.NewReader(body))
	return cloned, nil
}
