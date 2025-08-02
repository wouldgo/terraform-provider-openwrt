package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client interface {
	Auth(username, password string) error

	//UCI
	GetAll(section ...any) ([]System, error)
	GetSystem() (*System, error)
	TSet(data any, section ...any) error
	Add(section ...any) (string, error)
	Delete(section ...any) error
	CommitOrRevert(section ...any) []error

	//FS
	Writefile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	RemoveFile(path string) error

	//OPKG
	CheckPackage(pack string) (*PackageInfo, error)
	UpdatePackages() error
	InstallPackage(packages ...string) error
	RemovePackages(packages ...string) error
}

var _ Client = (*client)(nil)

type client struct {
	*uci
	*fs
	*opkg
	url string
}

func NewClient(url string) (Client, error) {
	if url == "" {
		return nil, fmt.Errorf("missing mandatory parameter url")
	}

	client := &client{
		url: url,
	}

	return client, nil
}

func (c *client) Auth(username, password string) error {
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
	resp, err := http.Post(u, "application/json; charset=utf-8", bytes.NewReader(b))
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

	c.uci = &uci{
		token: &data.Result,
		url:   &c.url,
	}
	c.fs = &fs{
		token: &data.Result,
		url:   &c.url,
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

func call(remoteUrl, token, rpc, method string, params []any) (json.RawMessage, error) {
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
	fmt.Println(string(b))
	resp, err := http.Post(u.String(), "application/json; charset=utf-8", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var responseBody jsonRPCResponseBody
	decoder := json.NewDecoder(resp.Body)
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
