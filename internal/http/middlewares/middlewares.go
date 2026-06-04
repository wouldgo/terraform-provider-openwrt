// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http_middlewares

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

var (
	ErrNotClonable = errors.New("value is not clonable")
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

func Clone[T *http.Request | *http.Response](val T) (T, error) {
	var zero T
	switch v := any(val).(type) {
	case *http.Request:
		req, err := CloneRequest(v)
		return any(req).(T), err
	case *http.Response:
		resp, err := cloneResponse(v)
		return any(resp).(T), err
	default:
		return zero, errors.Join(
			fmt.Errorf("%s", reflect.ValueOf(v).String()),
			ErrNotClonable,
		)
	}
}

func cloneResponse(resp *http.Response) (*http.Response, error) {
	if resp.Body == nil || resp.Body == http.NoBody {
		return resp, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	cloned := *resp // shallow copy of the struct
	cloned.Body = io.NopCloser(bytes.NewReader(body))
	return &cloned, nil
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
