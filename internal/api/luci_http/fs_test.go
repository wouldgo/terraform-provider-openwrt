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
	"slices"
	"testing"
	"time"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci_http"
	"github.com/foxboron/terraform-provider-openwrt/internal/api/testutil"
)

func TestFsRPCs(t *testing.T) {
	expectedHost,
		expectedToken,
		expectedUsername,
		expectedPassword := testutil.TestingData(t)
	expectedPath := "/test"
	exptectedFileContent := []byte{1, 2, 3}
	fileContentEncodedBytes, _ := json.Marshal(exptectedFileContent)
	expectedFileContentEncoded := string(fileContentEncodedBytes)
	expectedFileContentEncoded = expectedFileContentEncoded[1 : len(expectedFileContentEncoded)-1]
	mockedRoundTripper := fsMockedRoundTripper(
		t,
		expectedHost,
		expectedUsername,
		expectedPassword,
		expectedToken,
		expectedPath,
		expectedFileContentEncoded,
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

	fs, ok := (client).(luci.FsFacade)
	if !ok {
		t.Fatalf("expected a luci.FsFacade")
	}

	err = fs.Writefile(t.Context(), expectedPath, exptectedFileContent)
	if err != nil {
		t.Errorf("write file error: %v", err)
	}

	data, err := fs.ReadFile(
		t.Context(),
		expectedPath,
	)
	if err != nil {
		t.Errorf("read file error: %v", err)
	}
	if !slices.Equal(exptectedFileContent, data) {
		t.Errorf("read file read different file content. Expected %b - Returned %b", exptectedFileContent, data)
	}

	err = fs.RemoveFile(t.Context(), expectedPath)
	if err != nil {
		t.Errorf("write file error: %v", err)
	}
}

func fsMockedRoundTripper(
	t *testing.T,
	expectedHost *url.URL,
	username,
	password,
	authToken,
	path,
	fileContentEncoded string,
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
			fakeRawResult := json.RawMessage(`{"key": "value"}`)
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
			case "/cgi-bin/luci/rpc/fs":
				switch rpcReq.Method {
				case "writefile":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 2 {
						t.Errorf("expected len for params must be 2. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != path {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], path)
					}
					if rpcReq.Params[1] != fileContentEncoded {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[1], fileContentEncoded)
					}

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
				case "readfile":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != path {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], path)
					}

					raw := json.RawMessage("\"" + fileContentEncoded + "\"")

					rpcResp = testutil.RPCResponse{
						Result: raw,
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
				case "remove":
					paramsLen := len(rpcReq.Params)
					if paramsLen != 1 {
						t.Errorf("expected len for params must be 1. Found %d", paramsLen)
					}

					if rpcReq.Params[0] != path {
						t.Errorf("expected path param must be \"%s\". Found \"%s\"", rpcReq.Params[0], path)
					}

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
