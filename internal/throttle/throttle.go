// Package throttle provides rate-limiting for secret operations,
// ensuring providers are not overwhelmed by bursts of requests.
package throttle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Throttler limits the rate of operations using a token bucket approach.
type Throttler struct {
	mu       sync.Mutex
	tokens   map[string][]time.Time
	ratePerSec int
	window   time.Duration
}

// Option configures a Throttler.
type Option func(*Throttler)

// WithRate sets the maximum number of operations allowed per second per key.
func WithRate(ratePerSec int) Option {
	return func(t *Throttler) {
		if ratePerSec > 0 {
			t.ratePerSec = ratePerSec
		}
	}
}

// New creates a Throttler with the given options.
func New(opts ...Option) *Throttler {
	t := &Throttler{
		tokens:     make(map[string][]time.Time),
		ratePerSec: 10,
		window:     time.Second,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Allow reports whether an operation for the given key is permitted.
// It returns an error if the rate limit has been exceeded.
func (t *Throttler) Allow(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.window)

	// Evict timestamps outside the window.
	valid := t.tokens[key][:0]
	for _, ts := range t.tokens[key] {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	t.tokens[key] = valid

	if len(t.tokens[key]) >= t.ratePerSec {
		return fmt.Errorf("throttle: rate limit exceeded for key %q (%d ops/s)", key, t.ratePerSec)
	}

	t.tokens[key] = append(t.tokens[key], now)
	return nil
}

// Usage returns the current operation count within the window for a key.
func (t *Throttler) Usage(key string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.window)
	count := 0
	for _, ts := range t.tokens[key] {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded timestamps for a key.
func (t *Throttler) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tokens, key)
}
