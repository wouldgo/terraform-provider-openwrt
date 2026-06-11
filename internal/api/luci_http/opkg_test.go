// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci_http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci_http"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/testutil"
)

func TestOpkgRPCs(t *testing.T) {
	expectedHost,
		expectedToken,
		expectedUsername,
		expectedPassword := testutil.TestingData(t)
	expectedPackage := "test"
	expectedVersion := "1.1.1-test"
	expectedInstalled := true
	timeouts := newMockTimeouts(2 * time.Second)
	mockedRoundTripper := opkgMockedRoundTripper(
		t,
		expectedHost,
		expectedUsername,
		expectedPassword,
		expectedToken,
		expectedPackage,
		expectedVersion,
		expectedInstalled,
	)
	clientFactory, _ := luci_http.NewHTTPRPCFactory(luci_http.HTTPRPCConfiguration{
		RoundTripper: mockedRoundTripper,
	})

	client, err := clientFactory.Get(
		t.Context(),
		expectedHost.String(),
		expectedUsername,
		expectedPassword,
		timeouts,
	)
	if err != nil {
		t.Errorf("client get error %v", err)
	}

	opkg, ok := (client).(luci.OpkgFacade)
	if !ok {
		t.Fatalf("expected a luci.OpkgFacade")
	}
	err = opkg.UpdatePackages(
		t.Context(),
	)
	if err != nil {
		t.Errorf("update packages error: %v", err)
	}

	packageInfo, err := opkg.CheckPackage(
		t.Context(),
		expectedPackage,
	)
	if err != nil {
		t.Errorf("check package error: %v", err)
	}
	if packageInfo.Version != expectedVersion {
		t.Errorf("check package expected version not ok. Expected %s, Found %s", expectedVersion, packageInfo.Version)
	}
	if packageInfo.Status.Installed != expectedInstalled {
		t.Errorf("check package expected status installed not ok. Expected %t, Found %t", expectedInstalled, packageInfo.Status.Installed)
	}

	err = opkg.InstallPackages(
		t.Context(),
	)
	if err == nil || !errors.Is(err, luci.ErrPackagesNotSpecified) {
		t.Errorf("install no packages must return an error %s, got %s", luci.ErrPackagesNotSpecified, err)
	}

	err = opkg.InstallPackages(
		t.Context(),
		expectedPackage,
	)
	if err != nil {
		t.Errorf("install package error: %v", err)
	}

	err = opkg.RemovePackages(
		t.Context(),
	)
	if err == nil || !errors.Is(err, luci.ErrPackagesNotSpecified) {
		t.Errorf("remove no packages must return an error %s, got %s", luci.ErrPackagesNotSpecified, err)
	}

	err = opkg.RemovePackages(
		t.Context(),
		expectedPackage,
	)
	if err != nil {
		t.Errorf("remove package error: %v", err)
	}
}

func opkgMockedRoundTripper(
	t *testing.T,
	expectedHost *url.URL,
	username,
	password,
	authToken,
	expectedPackage,
	expectedVersion string,
	expectedInstalled bool,
) http.RoundTripper {
	return &testutil.MockRoundTripper{
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

			var (
				rpcReq  testutil.RPCRequest
				rpcResp testutil.RPCResponse
			)
			err = json.Unmarshal(data, &rpcReq)
			if err != nil {
				return nil, err
			}

			switch req.URL.Path {
			case "/cgi-bin/luci/rpc/auth":
				response, err := testutil.HandleAuth(authToken, username, password, req, data)
				if err != nil {
					t.Errorf("handling auth in error: %s", err.Error())
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(response)),
					Header:     make(http.Header),
				}, nil
			case "/cgi-bin/luci/rpc/ipkg":
				switch rpcReq.Method {
				case "update":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 0 {
						t.Errorf("expected len for params must be 0. Found %d", paramsLen)
					}

					fakeRawResult := json.RawMessage(`[0]`)
					rpcResp = testutil.RPCResponse{
						Result: fakeRawResult,
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
				case "status":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedPackage {
						t.Errorf("expected package must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedPackage)
					}

					result := json.RawMessage(`{
						"` + expectedPackage + `": {
							"Version": "` + expectedVersion + `",
							"Status": {
								"installed": ` + strconv.FormatBool(expectedInstalled) + `
							}
						}
					}`)
					rpcResp = testutil.RPCResponse{
						Result: result,
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
				case "install", "remove":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedPackage {
						t.Errorf("expected package must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedPackage)
					}

					fakeRawResult := json.RawMessage(`[0]`)
					rpcResp = testutil.RPCResponse{
						Result: fakeRawResult,
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
				default:
					t.Errorf("method \"%s\" not valid", rpcReq.Method)
					return nil, errors.New("method not valid")
				}

			default:
				t.Fatalf("path \"%s\" not permitted in this test", req.URL.Path)
				return nil, errors.New("path not permitted")
			}
		},
	}
}
