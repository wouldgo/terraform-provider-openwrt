// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build test

package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

type RPCRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  struct {
		Code    float64 `json:"code"`
		Message string  `json:"message"`
	} `json:"error"`
}

func MustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("URL error: %v", err)
	}
	return u
}

func TestingData(t *testing.T) (*url.URL, string, string, string) {
	expectedHost := MustURL(t, "http://127.0.0.1:8080")
	expectedToken := "token_123_star"
	expectedUsername := "root"
	expectedPassword := "password"

	return expectedHost, expectedToken, expectedUsername, expectedPassword
}

func SharedHTTPChecks(
	expectedHost *url.URL,
	authToken string,
	req *http.Request) error {
	requestURL := req.URL
	if requestURL.Host != expectedHost.Host {
		return fmt.Errorf("expected Host %s - arrived Host %s", expectedHost.Host, requestURL.Host)
	}

	if req.URL.Path != "/cgi-bin/luci/rpc/auth" &&
		req.URL.Query().Get("auth") != authToken {
		return errors.New("auth token must be present")
	}

	if req.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		return errors.New("header content-type has to be set to application/json; charset=utf-8")
	}
	return nil
}

func HandleAuth(
	authToken, username, password string,
	req *http.Request,
	data []byte) ([]byte, error) {
	if req.URL.Path == "/cgi-bin/luci/rpc/auth" {
		var loginData struct {
			Id     int      `json:"id"`
			Method string   `json:"method"`
			Params []string `json:"params"`
		}
		err := json.Unmarshal(data, &loginData)
		if err != nil {
			return nil, err
		}
		if loginData.Id != 1 {
			return nil, fmt.Errorf("id in auth is set to 1. Found %d", loginData.Id)
		}
		if loginData.Method != "login" {
			return nil, fmt.Errorf("method in auth is set to \"login\". Found %s", loginData.Method)
		}
		if !slices.Equal(loginData.Params, []string{username, password}) {
			return nil, fmt.Errorf("params in auth has to be set to \"%s\" username and \"%s\" password. Found \"%s\" username and \"%s\" password | [%v]",
				username, password, loginData.Params[0], loginData.Params[1], strings.Join(loginData.Params[2:], ", "))
		}

		responseData := struct {
			Result string `json:"result"`
			Error  string `json:"error"`
		}{
			Result: authToken,
			Error:  "",
		}

		response, err := json.Marshal(&responseData)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
	return nil, errors.New("not an auth")
}
