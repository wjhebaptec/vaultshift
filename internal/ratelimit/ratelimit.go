// Package ratelimit provides a token-bucket rate limiter for controlling
// the frequency of secret operations across providers.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter enforces a maximum number of operations per time window per key.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int
	window   time.Duration
	nowFn    func() time.Time
}

type bucket struct {
	tokens    int
	resetAt   time.Time
}

// Option configures a Limiter.
type Option func(*Limiter)

// WithRate sets the maximum number of allowed operations per window.
func WithRate(n int) Option {
	return func(l *Limiter) {
		if n > 0 {
			l.rate = n
		}
	}
}

// WithWindow sets the duration of each rate-limit window.
func WithWindow(d time.Duration) Option {
	return func(l *Limiter) {
		if d > 0 {
			l.window = d
		}
	}
}

// New creates a Limiter with the given options.
// Defaults: rate=10, window=1 minute.
func New(opts ...Option) *Limiter {
	l := &Limiter{
		buckets: make(map[string]*bucket),
		rate:    10,
		window:  time.Minute,
		nowFn:   time.Now,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Allow reports whether the operation identified by key is permitted.
// It returns an error if the rate limit has been exceeded.
func (l *Limiter) Allow(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFn()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{
			tokens:  l.rate - 1,
			resetAt: now.Add(l.window),
		}
		return nil
	}
	if b.tokens <= 0 {
		return fmt.Errorf("ratelimit: key %q exceeded %d ops per %s", key, l.rate, l.window)
	}
	b.tokens--
	return nil
}

// Remaining returns the number of tokens left for key in the current window.
func (l *Limiter) Remaining(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFn()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		return l.rate
	}
	return b.tokens
}

// Reset clears all tracked buckets.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]*bucket)
}
