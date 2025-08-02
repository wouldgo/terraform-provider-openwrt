package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type opkg struct {
	token  *string
	url    *string
	client *http.Client
}

type PackageInfo struct {
	Version string `json:"Version"`
	Status  Status `json:"Status"`
}

type Status struct {
	Installed bool `json:"installed"`
	// User      bool `json:"user"`
	// Install   bool `json:"install"`
}

func (c *opkg) UpdatePackages(ctx context.Context) error {
	result, err := call(ctx, c.client, *c.url, *c.token, "ipkg", "update", []any{})
	if err != nil {
		return err
	}

	var data []any
	if err = json.Unmarshal(result, &data); err != nil {
		return err
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return fmt.Errorf("update return value not a float64 value")
	}

	if retCasted != 0 {
		return fmt.Errorf("update packages returns %.0f", retCasted)
	}
	return nil
}

func (c *opkg) CheckPackage(ctx context.Context, pack string) (*PackageInfo, error) {
	result, err := call(ctx, c.client, *c.url, *c.token, "ipkg", "status", []any{pack})
	if err != nil {
		return nil, err
	}

	var data map[string]PackageInfo
	if err = json.Unmarshal(result, &data); err != nil {
		if err = json.Unmarshal(result, &[]bool{}); err != nil {

			return nil, err
		}

		return &PackageInfo{
			Version: "",
			Status: Status{
				Installed: false,
			},
		}, nil
	}
	ret, ok := data[pack]
	if !ok {
		return nil, fmt.Errorf("package not found")
	}
	return &ret, nil
}

func (c *opkg) InstallPackages(ctx context.Context, packages ...string) error {
	packagesLen := len(packages)
	if packagesLen == 0 {
		return fmt.Errorf("no packages specified")
	}

	toApi := make([]any, 0, packagesLen)
	for _, aPackage := range packages {
		toApi = append(toApi, aPackage)
	}
	result, err := call(ctx, c.client, *c.url, *c.token, "ipkg", "install", toApi)
	if err != nil {
		return err
	}

	var data []any
	if err = json.Unmarshal(result, &data); err != nil {
		return err
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return fmt.Errorf("install return value not a float64 value")
	}

	if retCasted != 0 {
		return fmt.Errorf("install packages returns %.0f", retCasted)
	}
	return nil
}

func (c *opkg) RemovePackages(ctx context.Context, packages ...string) error {
	packagesLen := len(packages)
	if packagesLen == 0 {
		return fmt.Errorf("no packages specified")
	}

	toApi := make([]any, 0, packagesLen)
	for _, aPackage := range packages {
		toApi = append(toApi, aPackage)
	}
	result, err := call(ctx, c.client, *c.url, *c.token, "ipkg", "remove", toApi)
	if err != nil {
		return err
	}

	var data []any
	if err = json.Unmarshal(result, &data); err != nil {
		return err
	}

	ret := data[0]
	retCasted, ok := ret.(float64)
	if !ok {
		return fmt.Errorf("remove return value not a float64 value")
	}

	if retCasted != 0 {
		return fmt.Errorf("remove packages returns %.0f", retCasted)
	}
	return nil
}
