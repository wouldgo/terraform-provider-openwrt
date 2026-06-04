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

func TestServiceRPCs(t *testing.T) {
	expectedHost,
		expectedToken,
		expectedUsername,
		expectedPassword := testutil.TestingData(t)
	expectedServiceName := "my_service"
	expectedServiceEnabled := true
	mockedRoundTripper := serviceHappyPathMockedRoundTripper(
		t,
		expectedHost,
		expectedUsername,
		expectedPassword,
		expectedToken,
		expectedServiceName,
		expectedServiceEnabled,
	)

	timeouts := newMockTimeouts(2 * time.Second)
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

	service, ok := (client).(luci.ServiceFacade)
	if !ok {
		t.Fatalf("expected a luci.ServiceFacade")
	}

	services, err := service.ListServices(
		t.Context(),
	)
	if err != nil {
		t.Errorf("list service error: %v", err)
	}

	servicesLen := len(services)
	if servicesLen != 1 {
		t.Errorf("expected services size is 1")
	}

	if services[0] != expectedServiceName {
		t.Errorf("expected service to be \"%s\", found \"%s\"", expectedServiceName, services[0])
	}

	isEnabled, err := service.IsEnabled(
		t.Context(),
		expectedServiceName,
	)
	if err != nil {
		t.Errorf("service is enabled error: %v", err)
	}

	if isEnabled != expectedServiceEnabled {
		t.Errorf("exptected service enabled to be \"%t\", found \"%t\"", expectedServiceEnabled, isEnabled)
	}

	err = service.DisableService(
		t.Context(),
		expectedServiceName,
	)
	if err != nil {
		t.Errorf("service disable service error: %v", err)
	}

	err = service.EnableService(
		t.Context(),
		expectedServiceName,
	)
	if err != nil {
		t.Errorf("service enable service error: %v", err)
	}

	err = service.StartService(
		t.Context(),
		expectedServiceName,
	)
	if err != nil {
		t.Errorf("service start service error: %v", err)
	}

	err = service.StopSevice(
		t.Context(),
		expectedServiceName,
	)
	if err != nil {
		t.Errorf("service stop service error: %v", err)
	}

	err = service.RestartService(
		t.Context(),
		expectedServiceName,
	)
	if err != nil {
		t.Errorf("restart service error: %v", err)
	}
}

func serviceHappyPathMockedRoundTripper(
	t *testing.T,
	expectedHost *url.URL,
	username,
	password,
	authToken,
	expectedServiceName string,
	expectedServiceEnabled bool,
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
			case "/cgi-bin/luci/rpc/sys":
				switch rpcReq.Method {
				case "init.names":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 0 {
						t.Errorf("expected len for params must be 0. Found %d", paramsLen)
					}

					fakeRawResult := json.RawMessage(`["` + expectedServiceName + `"]`)
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
				case "init.enabled":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedServiceName {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedServiceName)
					}

					fakeRawResult := json.RawMessage(strconv.FormatBool(expectedServiceEnabled))
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
				case "init.disable":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedServiceName {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedServiceName)
					}

					fakeRawResult := json.RawMessage(strconv.FormatBool(expectedServiceEnabled))
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
				case "init.enable":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedServiceName {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedServiceName)
					}

					fakeRawResult := json.RawMessage(strconv.FormatBool(expectedServiceEnabled))
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
				case "init.start":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedServiceName {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedServiceName)
					}

					fakeRawResult := json.RawMessage(strconv.FormatBool(expectedServiceEnabled))
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
				case "init.stop":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedServiceName {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedServiceName)
					}

					fakeRawResult := json.RawMessage(strconv.FormatBool(expectedServiceEnabled))
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
