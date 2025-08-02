package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client interface {
	Auth(ctx context.Context, username, password string) error

	//UCI
	GetAll(ctx context.Context, section ...any) ([]System, error)
	GetSystem(ctx context.Context) (*System, error)
	TSet(ctx context.Context, data any, section ...any) error
	Add(ctx context.Context, section ...any) (string, error)
	Delete(ctx context.Context, section ...any) error
	CommitOrRevert(ctx context.Context, section ...any) []error

	//FS
	Writefile(ctx context.Context, path string, data []byte) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	RemoveFile(ctx context.Context, path string) error

	//OPKG
	UpdatePackages(ctx context.Context) error
	CheckPackage(ctx context.Context, pack string) (*PackageInfo, error)
	InstallPackages(ctx context.Context, packages ...string) error
	RemovePackages(ctx context.Context, packages ...string) error
}

var _ Client = (*client)(nil)

type client struct {
	*uci
	*fs
	*opkg
	url    string
	client *http.Client
}

func NewClient(url string) (Client, error) {
	if url == "" {
		return nil, fmt.Errorf("missing mandatory parameter url")
	}

	client := &client{
		url:    url,
		client: &http.Client{},
	}

	return client, nil
}

func (c *client) Auth(ctx context.Context, username, password string) error {
	innerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	b, err := json.Marshal(&struct {
		Id     int      `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{
		Id:     1,
		Method: "login",
		Params: []string{username, password},
	})
	if err != nil {
		return err
	}
	u, err := url.JoinPath(c.url, "cgi-bin/luci/rpc/auth")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(innerCtx, http.MethodPost, u, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	var data struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to read authentication response body: %w", err)
	}

	if data.Error != "" {
		return fmt.Errorf("authentication error: %s", data.Error)
	}

	if data.Result == "" {
		return fmt.Errorf("empty reply as result")
	}

	c.uci = &uci{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}
	c.fs = &fs{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}
	c.opkg = &opkg{
		token:  &data.Result,
		url:    &c.url,
		client: c.client,
	}

	return nil
}

type jsonRPCRequestBody struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

var _ error = (*jsonRPCResponseError)(nil)

type jsonRPCResponseError struct {
	Code    float64 `json:"code"`
	Message string  `json:"message"`
}

func (j *jsonRPCResponseError) Error() string {
	return fmt.Sprintf("rpc call in error: %.0f: %s", j.Code, j.Message)
}

type jsonRPCResponseBody struct {
	Error  *jsonRPCResponseError `json:"error"`
	Result *json.RawMessage      `json:"result"`
}

func call(ctx context.Context, client *http.Client, remoteUrl, token, rpc, method string, params []any) (json.RawMessage, error) {
	innerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if remoteUrl == "" {
		return nil, fmt.Errorf("missing remote url")
	}
	if token == "" {
		return nil, fmt.Errorf("no auth is performed against %s", remoteUrl)
	}
	if rpc == "" {
		return nil, fmt.Errorf("missing rpc command to %s", remoteUrl)
	}
	if method == "" {
		return nil, fmt.Errorf("missing rpc command method: %s", rpc)
	}
	u, err := url.Parse(remoteUrl)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("cgi-bin/luci/rpc/%s", rpc)
	q := u.Query()
	q.Add("auth", token)
	u.RawQuery = q.Encode()

	b, err := json.Marshal(&jsonRPCRequestBody{
		Method: method,
		Params: params,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(innerCtx, http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	fmt.Printf("%s - %s: %s", req.Method, req.URL, string(b))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var responseBody jsonRPCResponseBody
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err = decoder.Decode(&responseBody); err != nil {
		return nil, err
	}
	if responseBody.Error != nil {
		return nil, responseBody.Error
	}

	if responseBody.Result == nil {
		return nil, fmt.Errorf("result is absent")
	}

	return *responseBody.Result, nil
}

// Purge the sections from the anonymous things
func purgeFields(d any) (any, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(b, &objmap); err != nil {
		return nil, err
	}
	delete(objmap, ".name")
	delete(objmap, ".anonymous")
	delete(objmap, ".type")
	return objmap, nil
}
