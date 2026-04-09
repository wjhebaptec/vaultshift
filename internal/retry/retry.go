// Package retry provides configurable retry logic with backoff strategies
// for use when interacting with remote secret manager providers.
package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// ErrMaxAttemptsReached is returned when all retry attempts are exhausted.
var ErrMaxAttemptsReached = errors.New("retry: max attempts reached")

// Config holds the configuration for retry behaviour.
type Config struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultConfig returns a sensible default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Retryer executes operations with retry logic.
type Retryer struct {
	cfg Config
}

// New creates a new Retryer with the provided Config.
func New(cfg Config) *Retryer {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 1.0
	}
	return &Retryer{cfg: cfg}
}

// Do executes fn, retrying on non-nil errors up to MaxAttempts times.
// It respects context cancellation between attempts.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if attempt < r.cfg.MaxAttempts-1 {
			delay := r.delay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return errors.Join(ErrMaxAttemptsReached, err)
}

// delay calculates the backoff duration for a given attempt index.
func (r *Retryer) delay(attempt int) time.Duration {
	d := float64(r.cfg.InitialDelay) * math.Pow(r.cfg.Multiplier, float64(attempt))
	if d > float64(r.cfg.MaxDelay) {
		d = float64(r.cfg.MaxDelay)
	}
	return time.Duration(d)
}
