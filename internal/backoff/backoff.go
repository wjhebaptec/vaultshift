// Package backoff provides configurable backoff strategies for retry logic.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy defines how delay is calculated between attempts.
type Strategy int

const (
	StrategyFixed      Strategy = iota // constant delay
	StrategyLinear                     // delay grows linearly
	StrategyExponential                // delay doubles each attempt
)

// Config holds backoff configuration.
type Config struct {
	Strategy   Strategy
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     bool
}

// DefaultConfig returns a sensible exponential backoff configuration.
func DefaultConfig() Config {
	return Config{
		Strategy:   StrategyExponential,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// Backoff computes the delay for a given attempt (0-indexed).
type Backoff struct {
	cfg Config
}

// New creates a new Backoff with the provided config.
func New(cfg Config) *Backoff {
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 100 * time.Millisecond
	}
	return &Backoff{cfg: cfg}
}

// Delay returns the wait duration before the given attempt number (1-indexed).
func (b *Backoff) Delay(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	var d time.Duration
	switch b.cfg.Strategy {
	case StrategyFixed:
		d = b.cfg.BaseDelay
	case StrategyLinear:
		d = b.cfg.BaseDelay * time.Duration(attempt)
	case StrategyExponential:
		factor := math.Pow(b.cfg.Multiplier, float64(attempt-1))
		d = time.Duration(float64(b.cfg.BaseDelay) * factor)
	}

	if b.cfg.MaxDelay > 0 && d > b.cfg.MaxDelay {
		d = b.cfg.MaxDelay
	}

	if b.cfg.Jitter && d > 0 {
		// add up to 20% random jitter
		jitter := time.Duration(rand.Int63n(int64(d / 5)))
		d += jitter
	}

	return d
}

// Reset is a no-op placeholder for stateful implementations.
func (b *Backoff) Reset() {}
