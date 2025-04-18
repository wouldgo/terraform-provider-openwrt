package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// func Encode[T any](w http.ResponseWriter, status int, v T) error {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)
// 	if err := json.NewEncoder(w).Encode(v); err != nil {
// 		return fmt.Errorf("encode json: %w", err)
// 	}
// 	return nil
// }
//
// func Decode[T any](b io.ReadCloser) (T, error) {
// 	var v T
// 	if err := json.NewDecoder(b).Decode(&v); err != nil {
// 		return v, fmt.Errorf("decode json: %w", err)
// 	}
// 	return v, nil
// }

func Unmarshal[T any](b json.RawMessage) (T, error) {
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// Merges a into b
func Merge[T any](a, b any) (T, error) {
	var v T
	var ma, mb map[string]json.RawMessage

	// Turn a,b any to bytes
	ba, err := json.Marshal(a)
	if err != nil {
		return *new(T), err
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return *new(T), err
	}

	// Turn ba and bb into our shimming map string -> raw message
	if err := json.Unmarshal(ba, &ma); err != nil {
		return *new(T), nil
	}
	if err := json.Unmarshal(bb, &mb); err != nil {
		return *new(T), nil
	}

	// Merge values from a into b
	// We could implement more logic here for special fields
	for k, v := range ma {
		mb[k] = v
		// switch k {
		// case ".name", ".anonymous", ".type":
		// 	continue
		// default:
		// 	mb[k] = v
		// }
	}

	// Marshal `mb` into T
	bf, err := json.Marshal(mb)
	if err != nil {
		return *new(T), err
	}
	if err := json.Unmarshal(bf, &v); err != nil {
		return *new(T), err
	}
	return v, nil
}

type Client struct {
	url   string
	token string
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) Auth(username, password string) error {
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
		log.Fatalf("Failed to read response body: %v", err)
	}
	c.token = data.Result
	return nil
}

type jsonRPCResponseBody struct {
	Error  *string          `json:"error"`
	Result *json.RawMessage `json:"result"`
}

func (c *Client) uciCall(rpc string, method string, params []any) (*json.RawMessage, error) {
	u, err := url.Parse(c.url)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("cgi-bin/luci/rpc/%s", rpc)
	q := u.Query()
	q.Add("auth", c.token)
	u.RawQuery = q.Encode()

	b, err := json.Marshal(&struct {
		Method string `json:"method"`
		Params []any  `json:"params"`
	}{
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
		return nil, fmt.Errorf(*responseBody.Error)
	}
	return responseBody.Result, nil
}

func (c *Client) UCIGetAll(section ...any) (*json.RawMessage, error) {
	return c.uciCall("uci", "get_all", section)
}

func UCIGetAllT[T any](c *Client, section ...any) (T, error) {
	r, err := c.uciCall("uci", "get_all", section)
	if err != nil {
		return *new(T), err
	}
	return Unmarshal[T](*r)
}

// Purge the sections from the anonymous things
func PurgeFields(d any) (any, error) {
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

func (c *Client) UCITSet(data any, section ...any) (*json.RawMessage, error) {
	data, err := PurgeFields(&data)
	if err != nil {
		return nil, err
	}
	section = append(section, data)
	return c.uciCall("uci", "tset", section)
}

func (c *Client) UCISection(data any, section ...any) (*json.RawMessage, error) {
	data, err := PurgeFields(&data)
	if err != nil {
		return nil, err
	}
	section = append(section, data)
	return c.uciCall("uci", "section", section)
}

func (c *Client) UCIAdd(section ...any) (string, error) {
	raw, err := c.uciCall("uci", "add", section)
	if err != nil {
		return "", err
	}
	var s string
	if err := json.Unmarshal(*raw, &s); err != nil {
		return "", err
	}
	return s, nil
}

func (c *Client) UCICommit(section ...any) (*json.RawMessage, error) {
	return c.uciCall("uci", "commit", section)
}

func (c *Client) UCIRevert(section ...any) (*json.RawMessage, error) {
	return c.uciCall("uci", "revert", section)
}

func (c *Client) UCICommitAndRevert(section ...any) (*json.RawMessage, diag.Diagnostics) {
	var diag diag.Diagnostics
	r, err := c.UCICommit(section...)
	if err != nil {
		diag.AddError(fmt.Sprintf("Failed to update configu %q", section), err.Error())
		_, err = c.UCIRevert(section...)
		if err != nil {
			diag.AddError(fmt.Sprintf("Failed to revert configu %q", section), err.Error())
			return nil, diag
		}
	}
	return r, diag
}

func (c *Client) UCIDelete(section ...any) (*json.RawMessage, error) {
	return c.uciCall("uci", "delete", section)
}

func (c *Client) Writefile(path string, data []byte) (*json.RawMessage, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return c.uciCall("fs", "writefile", []any{path, data})
}

func (c *Client) ReadFile(path string) ([]byte, error) {
	raw, err := c.uciCall("fs", "readfile", []any{path})
	if err != nil {
		return nil, err
	}
	var s string
	if err := json.Unmarshal(*raw, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

func (c *Client) RemoveFile(path string) (*json.RawMessage, error) {
	return c.uciCall("fs", "remove", []any{path})
}

type AnonymousSection struct {
	Name      string `json:".name"`
	Anonymous bool   `json:".anonymous"`
	Type      string `json:".type"`
}

func IsAnonymousSection(raw json.RawMessage) bool {
	var anon AnonymousSection
	if err := json.Unmarshal(raw, &anon); err != nil {
		panic(err)
	}
	return anon.Anonymous == true && anon.Type == "system"
}

func (c *Client) UCIGetSystem() (json.RawMessage, error) {
	raw, err := c.UCIGetAll("system")
	if err != nil {
		return nil, err
	}
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(*raw, &objmap); err != nil {
		panic(err)
	}
	for k, v := range objmap {
		// We are partially assuming that the system settings is the first anonymous section
		if !strings.Contains(k, "cfg") {
			continue
		}
		if IsAnonymousSection(v) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("did not find the system section")
}
