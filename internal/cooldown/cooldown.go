// Package cooldown enforces a minimum wait period between repeated operations
// on the same key, preventing rapid re-execution of sensitive actions such as
// secret rotation or propagation.
package cooldown

import (
	"errors"
	"sync"
	"time"
)

// ErrCooldownActive is returned when an operation is attempted before the
// cooldown period for that key has elapsed.
var ErrCooldownActive = errors.New("cooldown: operation not permitted, cooldown period is active")

// Option configures a Cooldown.
type Option func(*Cooldown)

// WithClock overrides the time source used by the Cooldown (useful in tests).
func WithClock(fn func() time.Time) Option {
	return func(c *Cooldown) { c.now = fn }
}

// Cooldown tracks the last-allowed time for each key and rejects calls that
// arrive before the configured period has elapsed.
type Cooldown struct {
	period time.Duration
	now    func() time.Time
	mu     sync.Mutex
	last   map[string]time.Time
}

// New creates a Cooldown with the given minimum period between allowed calls.
func New(period time.Duration, opts ...Option) (*Cooldown, error) {
	if period <= 0 {
		return nil, errors.New("cooldown: period must be positive")
	}
	c := &Cooldown{
		period: period,
		now:    time.Now,
		last:   make(map[string]time.Time),
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Allow returns nil if the key is not in cooldown, recording the current time
// as the last-allowed instant. It returns ErrCooldownActive otherwise.
func (c *Cooldown) Allow(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if t, ok := c.last[key]; ok && now.Sub(t) < c.period {
		return ErrCooldownActive
	}
	c.last[key] = now
	return nil
}

// Remaining returns how long until the cooldown for key expires.
// It returns 0 if the key is not in cooldown.
func (c *Cooldown) Remaining(key string) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.last[key]
	if !ok {
		return 0
	}
	remaining := c.period - c.now().Sub(t)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset clears the cooldown record for a key, allowing it to proceed immediately.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, key)
}
