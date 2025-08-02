package api

import (
	"encoding/base64"
	"encoding/json"
)

type fs struct {
	token *string
	url   *string
}

func (c *fs) Writefile(path string, data []byte) error {
	_, err := call(*c.url, *c.token, "fs", "writefile", []any{path, data})
	return err
}

func (c *fs) ReadFile(path string) ([]byte, error) {
	raw, err := call(*c.url, *c.token, "fs", "readfile", []any{path})
	if err != nil {
		return nil, err
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(s)
}

func (c *fs) RemoveFile(path string) error {
	_, err := call(*c.url, *c.token, "fs", "remove", []any{path})
	return err
}
