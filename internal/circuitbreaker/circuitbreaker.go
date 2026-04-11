// Package circuitbreaker implements a simple circuit breaker pattern
// to prevent cascading failures when interacting with secret providers.
package circuitbreaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking calls
	StateHalfOpen              // testing recovery
)

var ErrOpen = errors.New("circuit breaker is open")

// Config holds configuration for the circuit breaker.
type Config struct {
	MaxFailures  int
	ResetTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
	}
}

// CircuitBreaker tracks failures and opens the circuit when the threshold is exceeded.
type CircuitBreaker struct {
	mu           sync.Mutex
	cfg          Config
	state        State
	failureCount int
	lastFailure  time.Time
}

// New creates a new CircuitBreaker with the given config.
func New(cfg Config) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = DefaultConfig().MaxFailures
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = DefaultConfig().ResetTimeout
	}
	return &CircuitBreaker{cfg: cfg}
}

// Allow returns nil if the call is permitted, or ErrOpen if the circuit is open.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.cfg.ResetTimeout {
			cb.state = StateHalfOpen
			return nil
		}
		return fmt.Errorf("%w: retry after %s", ErrOpen, cb.cfg.ResetTimeout)
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess resets the circuit breaker to closed state.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = StateClosed
}

// RecordFailure increments the failure count and opens the circuit if the threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailure = time.Now()
	if cb.failureCount >= cb.cfg.MaxFailures {
		cb.state = StateOpen
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Failures returns the current failure count.
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failureCount
}
