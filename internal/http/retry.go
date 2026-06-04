// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package http

import (
	"context"
	"math"
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
	// MinBackOff is the minimum backoff duration, used as a floor to avoid
	// request spam regardless of the computed value.
	MinBackOff time.Duration
	// MinRetries is the target number of retries during the early phase
	// (progress <= FadeEnd). Lower values produce larger backoffs (sparser retries).
	MinRetries uint8
	// MaxRetries is the target number of retries during the late phase
	// (progress > FadeEnd). Lower values produce larger backoffs (sparser retries).
	// Should be >= MinRetries.
	MaxRetries uint8
	// [0, 0.6] progress threshold after which determine the phase
	FadeEnd float64
}

// NextBackoffDuration computes the next backoff duration by dividing the
// remaining time budget across the interpolated number of retries, minus
// the average IO time.
//
// The backoff formula is:
//
// backoff = (remaining / nRetries) - avgIO
//
// where nRetries is interpolated between MinRetries and MaxRetries using an
// ease-out quadratic curve that activates after <FadeEnd>% progress.
//
// The returned duration is always bounded below by MinBackOff.
//
// Special cases:
//   - timeout <= 0: returns 0 immediately.
//   - elapsed >= timeout: returns 0 (no budget left).
//   - avgIO <= 0: defaults to MinBackOff as a baseline IO estimate.
func (mm MinMaxRetriesBackOffStrategy) NextBackoffDuration(ctx context.Context, elapsed, timeout, avgIO time.Duration) time.Duration {
	if timeout <= 0 {
		return 0
	}

	if avgIO <= 0 {
		avgIO = mm.MinBackOff
	}

	fadeEnd := clamp(mm.FadeEnd, 0, 0.6)
	remaining := timeout - elapsed
	if remaining <= 0 {
		return 0
	}

	progress := float64(elapsed) / float64(timeout)

	minRetries := float64(mm.MinRetries)
	maxRetries := math.Max(float64(mm.MaxRetries), minRetries)

	var ease float64
	switch {
	case progress <= fadeEnd:
		ease = 0
	case progress >= 1:
		ease = 1
	default:
		x := (progress - fadeEnd)
		ease = 1 - ((1 - x) * (1 - x))
	}

	nRetries := minRetries + (maxRetries-minRetries)*ease
	if nRetries <= 0 {
		nRetries = 1
	}

	backoff := time.Duration(float64(remaining)/nRetries) - avgIO

	if backoff < mm.MinBackOff {
		backoff = mm.MinBackOff
	}

	tflog.Debug(ctx, "next backoff duration", map[string]any{
		"timeout":   timeout.String(),
		"elapsed":   elapsed.String(),
		"remaining": remaining.String(),
		"avgIO":     avgIO.String(),
		"progress":  progress,
		"backoff":   backoff.String(),
	})

	return backoff
}

type AggressivenessBackOffStrategy struct {
	// MinBackOff is the minimum backoff duration, used as a floor to avoid
	// request spam regardless of the computed value.
	MinBackOff time.Duration
	// [0, 1] value to evaluate how aggressive retries has to be when > <FadeEnd>% progress is reached.
	// 0 very conservative (high backoff, sparse retries).
	// 1 very aggressive (low backoff, dense retries).
	Aggressiveness float64
	// [0, 0.8] progress threshold after which the "all-in" phase begins.
	// Below this value, backoff is modulated by Aggressiveness.
	// Above this value, backoff decreases quadratically regardless of Aggressiveness.
	// Defaults to 0 if not set (full all-in phase).
	FadeEnd float64
}

// NextBackoffDuration computes the next backoff duration for a retry attempt.
// It has two approaches: conservative and aggressive retry strategies based on how
// much of the timeout has elapsed (progress).
//
// The backoff is calculated in two phases, separated by the FadeEnd threshold:
//
// Phase 1 (progress < FadeEnd):
//
// The backoff is modulated by Aggressiveness. The closer progress is to
// FadeEnd, the smaller the backoff becomes (quadratic decay toward zero).
//   - Aggressiveness = 0: high backoff, sparse retries.
//   - Aggressiveness = 1: low backoff, dense retries.
//
// Phase 2 (progress >= FadeEnd, "all-in"):
//
// Aggressiveness is no longer considered. The backoff starts from a
// peak of (1 - Aggressiveness) and decays quadratically to zero as progress
// approaches 1. This for making it increasingly dense near the end
// of the timeout, regardless of the configured aggressiveness.
//
// The returned duration is always bounded:
//   - Never exceeds the remaining time budget (hard cap).
//   - Never goes below MinBackOff (soft floor, to avoid request spam).
//
// Special cases:
//   - timeout <= 0: returns 0 immediately.
//   - elapsed >= timeout: returns 0 (no budget left).
//   - avgIO <= 0: defaults to MinBackOff as a baseline IO estimate.
func (a AggressivenessBackOffStrategy) NextBackoffDuration(ctx context.Context, elapsed time.Duration, timeout time.Duration, avgIO time.Duration) time.Duration {
	if timeout <= 0 {
		return 0
	}

	if avgIO <= 0 {
		avgIO = a.MinBackOff
	}

	fadeEnd := clamp(a.FadeEnd, 0, 0.8)
	aggressiveness := clamp(a.Aggressiveness, 0, 1)

	remaining := timeout - elapsed
	if remaining <= 0 {
		return 0
	}

	progress := float64(elapsed) / float64(timeout)

	var K float64
	switch {
	case progress < fadeEnd:
		x := (progress - fadeEnd) / (1 - progress)
		f := x * x
		K = f * (1 - aggressiveness)
	default:
		peak := 1 - aggressiveness
		t := (progress - fadeEnd) / (1 - fadeEnd)
		K = peak * (1 - t) * (1 - t)
	}

	backoff := time.Duration(float64(avgIO) * K)

	// hard safety: never exceed remaining budget
	if backoff > remaining {
		backoff = remaining
	}

	// soft safety floor (avoid spam)
	if backoff < a.MinBackOff {
		backoff = a.MinBackOff
	}

	tflog.Debug(ctx, "next backoff duration", map[string]any{
		"timeout":   timeout.String(),
		"elapsed":   elapsed.String(),
		"remaining": remaining.String(),
		"avgIO":     avgIO.String(),
		"progress":  progress,
		"backoff":   backoff.String(),
	})

	return backoff
}

func clamp(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
