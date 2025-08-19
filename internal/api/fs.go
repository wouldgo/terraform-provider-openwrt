package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type fs struct {
	token  *string
	url    *string
	client *http.Client
}

func (c *fs) Writefile(ctx context.Context, path string, data []byte) error {
	_, err := call(ctx, c.client, *c.url, *c.token, "fs", "writefile", []any{path, data})
	return err
}

func (c *fs) ReadFile(ctx context.Context, path string) ([]byte, error) {
	raw, err := call(ctx, c.client, *c.url, *c.token, "fs", "readfile", []any{path})
	if err != nil {
		return nil, err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

func (c *fs) RemoveFile(ctx context.Context, path string) error {
	_, err := call(ctx, c.client, *c.url, *c.token, "fs", "remove", []any{path})
	return err
}
