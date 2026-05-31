// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ BackoffDurationStrategy = MinMaxRetriesBackOffStrategy{}
	_ BackoffDurationStrategy = AggressivenessBackOffStrategy{}
)

type BackoffDurationStrategy interface {
	NextBackoffDuration(ctx context.Context, elapsed, timeout, avgIO time.Duration) time.Duration
}

type MinMaxRetriesBackOffStrategy struct {
	//min amount of backoff to let openwrt not dying
	MinBackOff time.Duration
	//decreasing will increase the backoff time when <= 60% progress is reached
	MinRetries int8
	//decreasing will increase the backoff time when > 60% progress is reached
	MaxRetries int8
}

func (mm MinMaxRetriesBackOffStrategy) NextBackoffDuration(ctx context.Context, elapsed, timeout, avgIO time.Duration) time.Duration {
	if timeout <= 0 {
		return 0
	}

	if avgIO <= 0 {
		avgIO = 100 * time.Millisecond
	}

	remaining := timeout - elapsed
	if remaining <= 0 {
		return 0
	}

	progress := float64(elapsed) / float64(timeout)

	minRetries := float64(mm.MinRetries)
	maxRetries := float64(mm.MaxRetries)

	var ease float64
	switch {
	case progress <= 0.6:
		ease = 0
	case progress >= 1:
		ease = 1
	default:
		x := (progress - 0.6) / 0.4
		ease = 1 - ((1 - x) * (1 - x))
	}

	nRetries := minRetries + (maxRetries-minRetries)*ease

	backoff := time.Duration(float64(remaining)/nRetries) - avgIO

	if backoff < mm.MinBackOff {
		backoff = mm.MinBackOff
	}

	tflog.Debug(ctx, "next backoff duration", map[string]any{
		"timeout":  timeout.String(),
		"elapsed":  elapsed.String(),
		"remaning": remaining.String(),
		"avgIO":    avgIO.String(),
		"progress": progress,
		"backoff":  backoff.String(),
	})

	fmt.Printf("timeout %s, elapsed %s, remaning %s, avgIO %s, progress %f, backoff %s\r\n",
		timeout, elapsed, remaining, avgIO, progress, backoff)
	return backoff
}

type AggressivenessBackOffStrategy struct {
	//min amount of backoff to let openwrt not dying
	MinBackOff time.Duration
	//[0, 1] value to evaluate how aggressive retries has to be when > 60% progress is reached.
	// 0 very conservative (min amount of retries).
	// 1 very aggressive (a lot of retries).
	Aggressiveness float64
}

func (a AggressivenessBackOffStrategy) NextBackoffDuration(ctx context.Context, elapsed time.Duration, timeout time.Duration, avgIO time.Duration) time.Duration {
	if timeout <= 0 {
		return 0
	}

	if avgIO <= 0 {
		avgIO = 100 * time.Millisecond
	}

	aggressiveness := a.Aggressiveness
	if a.Aggressiveness < 0 {
		aggressiveness = 0
	}
	if a.Aggressiveness > 1 {
		aggressiveness = 1
	}

	remaining := timeout - elapsed
	if remaining <= 0 {
		return 0
	}

	progress := float64(elapsed) / float64(timeout)

	var f float64
	switch {
	case progress <= 0.6:
		f = 0
	case progress >= 1:
		f = 1
	default:
		x := (progress - 0.6) / 0.4
		f = x * x
	}

	A := 5.0 * aggressiveness
	K := 1 + A*f

	backoff := time.Duration(float64(avgIO) * (K - 1))

	// hard safety: never exceed remaining budget
	if backoff > remaining {
		backoff = remaining
	}

	// soft safety floor (avoid spam)
	if backoff < a.MinBackOff {
		backoff = a.MinBackOff
	}

	tflog.Debug(ctx, "next backoff duration", map[string]any{
		"timeout":  timeout.String(),
		"elapsed":  elapsed.String(),
		"remaning": remaining.String(),
		"avgIO":    avgIO.String(),
		"progress": progress,
		"backoff":  backoff.String(),
	})

	fmt.Printf("timeout %s, elapsed %s, remaning %s, avgIO %s, progress %f, backoff %s\r\n",
		timeout, elapsed, remaining, avgIO, progress, backoff)
	return backoff
}
