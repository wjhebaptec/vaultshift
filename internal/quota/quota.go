// Package quota provides rate-limiting and usage-quota enforcement
// for secret operations across providers.
package quota

import (
	"errors"
	"sync"
	"time"
)

// ErrQuotaExceeded is returned when an operation exceeds its allowed quota.
var ErrQuotaExceeded = errors.New("quota: operation limit exceeded")

// Entry tracks usage for a single quota bucket.
type Entry struct {
	Count     int
	WindowEnd time.Time
}

// Option configures a Limiter.
type Option func(*Limiter)

// WithWindow sets the rolling window duration.
func WithWindow(d time.Duration) Option {
	return func(l *Limiter) { l.window = d }
}

// WithLimit sets the maximum number of operations per window.
func WithLimit(n int) Option {
	return func(l *Limiter) { l.limit = n }
}

// Limiter enforces per-key operation quotas within a rolling time window.
type Limiter struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	bucket map[string]*Entry
}

// New creates a Limiter with the given options.
// Defaults: window = 1 minute, limit = 60.
func New(opts ...Option) *Limiter {
	l := &Limiter{
		window: time.Minute,
		limit:  60,
		bucket: make(map[string]*Entry),
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Allow checks whether the given key is within quota and increments its counter.
// Returns ErrQuotaExceeded if the limit has been reached for the current window.
func (l *Limiter) Allow(key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.bucket[key]
	if !ok || now.After(e.WindowEnd) {
		l.bucket[key] = &Entry{Count: 1, WindowEnd: now.Add(l.window)}
		return nil
	}
	if e.Count >= l.limit {
		return ErrQuotaExceeded
	}
	e.Count++
	return nil
}

// Usage returns the current count and window-end time for a key.
// If the window has expired or the key is unknown, count is 0.
func (l *Limiter) Usage(key string) (count int, windowEnd time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.bucket[key]
	if !ok || time.Now().After(e.WindowEnd) {
		return 0, time.Time{}
	}
	return e.Count, e.WindowEnd
}

// Reset clears the quota entry for a key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.bucket, key)
}
