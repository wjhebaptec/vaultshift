// Package cutoff provides a time-based cutoff gate that blocks operations
// on secrets that were last modified before a configured threshold.
package cutoff

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrBeforeCutoff is returned when a secret's timestamp precedes the cutoff.
var ErrBeforeCutoff = errors.New("cutoff: value predates cutoff threshold")

// Clock returns the current time. Replaceable for testing.
type Clock func() time.Time

// Cutoff gates operations based on a rolling time window.
type Cutoff struct {
	mu       sync.RWMutex
	window   time.Duration
	clock    Clock
	records  map[string]time.Time
}

// Option configures a Cutoff.
type Option func(*Cutoff)

// WithClock overrides the default clock.
func WithClock(c Clock) Option {
	return func(g *Cutoff) { g.clock = c }
}

// New creates a Cutoff that rejects values older than window.
func New(window time.Duration, opts ...Option) (*Cutoff, error) {
	if window <= 0 {
		return nil, errors.New("cutoff: window must be positive")
	}
	g := &Cutoff{
		window:  window,
		clock:   time.Now,
		records: make(map[string]time.Time),
	}
	for _, o := range opts {
		o(g)
	}
	return g, nil
}

// Mark records the timestamp for a secret key.
func (g *Cutoff) Mark(key string, at time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.records[key] = at
}

// Allow returns nil if the key's recorded time is within the window,
// or ErrBeforeCutoff if it predates the threshold.
func (g *Cutoff) Allow(key string) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	at, ok := g.records[key]
	if !ok {
		return fmt.Errorf("cutoff: no record for key %q", key)
	}
	threshold := g.clock().Add(-g.window)
	if at.Before(threshold) {
		return ErrBeforeCutoff
	}
	return nil
}

// Forget removes the record for a key.
func (g *Cutoff) Forget(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.records, key)
}
