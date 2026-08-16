// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http_transport

import (
	"net/http"
	"slices"
	"sync"
)

var (
	_ http.RoundTripper = (*DynamicTransport)(nil)
)

type RoundTripperFunc func(http.RoundTripper) http.RoundTripper
type namedRoundTripper struct {
	name string
	fn   RoundTripperFunc
}

type DynamicTransport struct {
	mu            sync.RWMutex
	base          http.RoundTripper
	roundtrippers []namedRoundTripper
	chain         http.RoundTripper // rebuilt on every mutation
}

func NewDynamicTransport(base http.RoundTripper) (*DynamicTransport, error) {
	if base == nil {
		base = http.DefaultTransport
	}
	dt := &DynamicTransport{
		base: base,
	}
	dt.chain = base
	return dt, nil
}

func (dt *DynamicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dt.mu.RLock()
	chain := dt.chain
	dt.mu.RUnlock()
	return chain.RoundTrip(req)
}

// must be called with dt.mu held (write lock)
func (dt *DynamicTransport) rebuild() {
	chain := dt.base
	for i := len(dt.roundtrippers) - 1; i >= 0; i-- {
		chain = dt.roundtrippers[i].fn(chain)
	}
	dt.chain = chain
}

func (dt *DynamicTransport) Add(name string, mw RoundTripperFunc) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	// replace func if name already exists
	for i, m := range dt.roundtrippers {
		if m.name == name {
			dt.roundtrippers[i].fn = mw
			dt.rebuild()
			return
		}
	}
	dt.roundtrippers = append(dt.roundtrippers, namedRoundTripper{
		name,
		mw,
	})
	dt.rebuild()
}

func (dt *DynamicTransport) Remove(name string) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	for i, m := range dt.roundtrippers {
		if m.name == name {
			dt.roundtrippers = slices.Delete(dt.roundtrippers, i, i+1)
			dt.rebuild()
			return
		}
	}
}
