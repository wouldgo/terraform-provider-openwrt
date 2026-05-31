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
	"testing"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci_http"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/testutil"
)

func TestSystemRPCs(t *testing.T) {
	expectedHost,
		expectedToken,
		expectedUsername,
		expectedPassword := testutil.TestingData(t)
	expectedSection := "system"
	mockedRoundTripper := systemHappyPathMockedRoundTripper(
		t,
		expectedHost,
		expectedUsername,
		expectedPassword,
		expectedToken,
		expectedSection,
	)
	timeouts := newMockTimeouts(2 * time.Second)
	clientFactory, _ := luci_http.NewHTTPRPCFactory(mockedRoundTripper, nil)

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

	system, ok := (client).(luci.SystemFacade)
	if !ok {
		t.Fatalf("expected a luci.SystemFacade")
	}

	systems, err := system.GetAll(
		t.Context(),
		expectedSection,
	)
	if err != nil {
		t.Errorf("get all error: %v", err)
	}
	if len(systems) != 1 {
		t.Errorf("expected system length is 1")
	}

	_, err = system.GetSystem(
		t.Context(),
	)
	if err != nil {
		t.Errorf("get system error: %v", err)
	}

	err = system.TSet(
		t.Context(),
		luci.System{},
		expectedSection,
		expectedSection,
	)
	if err != nil {
		t.Errorf("tset system error: %v", err)
	}

	str, err := system.Add(
		t.Context(),
		expectedSection,
	)
	if err != nil {
		t.Errorf("add system error: %v", err)
	}
	if str == "" {
		t.Errorf("add system string not ok.")
	}

	err = system.Delete(
		t.Context(),
		expectedSection,
	)
	if err != nil {
		t.Errorf("delete system error: %v", err)
	}

	err = system.CommitOrRevert(
		t.Context(),
		expectedSection,
	)
	if err != nil {
		t.Errorf("commit or revert error: %v", err)
	}
}

func systemHappyPathMockedRoundTripper(
	t *testing.T,
	expectedHost *url.URL,
	username,
	password,
	authToken,
	expectedSection string,
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
			case "/cgi-bin/luci/rpc/uci":
				switch rpcReq.Method {
				case "get_all":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedSection {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedSection)
					}

					fakeRawResult := json.RawMessage(`{
						"` + expectedSection + `" : {
							".name": "the_name",
							".type": "system",
							".anonymous": true,
							"hostname": "the_hostname",
							"description": "the_description",
							"notes": "the_notes",
							"buffersize": "the_buffersize",
							"conloglevel": "the_conloglevel",
							"cronloglevel": "the_cronloglevel",
							"klogconloglevel": "the_klogconloglevel",
							"log_buffer_size": "the_log_buffer_size",
							"log_file": "the_log_file",
							"log_hostname": "the_log_hostname",
							"log_ip": "the_log_ip",
							"log_port": "the_log_port",
							"log_prefix": "the_log_prefix",
							"log_proto": "the_log_proto",
							"log_remote": "the_log_remote",
							"log_size": "the_log_size",
							"log_trailer_null": "the_log_trailer_null",
							"log_type": "the_log_type",
							"ttylogin": "the_ttylogin",
							"urandom_seed": "the_urandom_seed",
							"timezone": "the_timezone",
							"zonename": "the_zonename",
							"zram_comp_algo": "the_zram_comp_algo",
							"zram_size_mb": "the_zram_size_mb"
						}
					}`)
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
				case "tset":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 3 {
						t.Errorf("expected len for params must be 3. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedSection {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedSection)
					}
					if rpcReq.Params[1] != expectedSection {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[1], expectedSection)
					}

					fakeRawResult := json.RawMessage(`{"key": "value"}`)

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
				case "add":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedSection {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedSection)
					}

					fakeRawResult := json.RawMessage(`"OK"`)

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
				case "delete":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedSection {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedSection)
					}

					fakeRawResult := json.RawMessage(`"OK"`)

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
				case "commit", "revert":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != expectedSection {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], expectedSection)
					}

					fakeRawResult := json.RawMessage(`true`)

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
