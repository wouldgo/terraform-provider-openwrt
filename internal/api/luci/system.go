// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package luci

import (
	"context"
	"time"
)

type SystemFacade interface {
	GetAll(ctx context.Context, section ...any) ([]System, error)
	GetSystem(ctx context.Context) (System, error)
	TSet(ctx context.Context, data any, section ...any) error
	Add(ctx context.Context, section ...any) (string, error)
	Delete(ctx context.Context, section ...any) error
	CommitOrRevert(ctx context.Context, section ...any) error
}

type SystemTimeouts interface {
	GetAll() time.Duration
	TSet() time.Duration
	Add() time.Duration
	Delete() time.Duration
	CommitOrRevert() time.Duration
}

type System struct {
	Id        string `json:".name,omitempty"`
	Type      string `json:".type,omitzero,omitempty"`
	Anonymous bool   `json:".anonymous,omitzero,omitempty"`

	Hostname        string `json:"hostname,omitzero"`
	Description     string `json:"description,omitzero"`
	Notes           string `json:"notes,omitzero"`
	Buffersize      string `json:"buffersize,omitzero"`
	ConLogLevel     string `json:"conloglevel,omitzero"`
	CronLogLevel    string `json:"cronloglevel,omitzero"`
	KlogconLogLevel string `json:"klogconloglevel,omitzero"`
	LogBufferSize   string `json:"log_buffer_size,omitzero"`
	LogFile         string `json:"log_file,omitzero"`
	LogHostname     string `json:"log_hostname,omitzero"`
	LogIP           string `json:"log_ip,omitzero"`
	LogPort         string `json:"log_port,omitzero"`
	LogPrefix       string `json:"log_prefix,omitzero"`
	LogProto        string `json:"log_proto,omitzero"`
	LogRemote       string `json:"log_remote,omitzero"`
	LogSize         string `json:"log_size,omitzero"`
	LogTrailerNull  string `json:"log_trailer_null,omitzero"`
	LogType         string `json:"log_type,omitzero"`
	TTYLogin        string `json:"ttylogin,omitzero"`
	UrandomSeed     string `json:"urandom_seed,omitempty"`
	Timezone        string `json:"timezone,omitzero"`
	ZoneName        string `json:"zonename,omitzero"`
	ZramCompAlgo    string `json:"zram_comp_algo,omitzero"`
	ZramSizeMb      string `json:"zram_size_mb,omitzero"`
}
