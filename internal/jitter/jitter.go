// Package jitter adds randomised delay to secret operations to avoid
// thundering-herd problems when many workers rotate secrets simultaneously.
package jitter

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Provider is the subset of the secret-manager interface used by jitter.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]string, error)
}

// Jitter wraps a Provider and sleeps for a random duration in [0, max) before
// each mutating operation (Put / Delete).
type Jitter struct {
	inner  Provider
	max    time.Duration
	sleep  func(context.Context, time.Duration) error
}

// Option configures a Jitter instance.
type Option func(*Jitter)

// WithSleepFunc replaces the default sleep implementation (useful in tests).
func WithSleepFunc(fn func(context.Context, time.Duration) error) Option {
	return func(j *Jitter) { j.sleep = fn }
}

// New creates a Jitter wrapper around inner. max is the upper bound of the
// random delay applied before mutating operations.
func New(inner Provider, max time.Duration, opts ...Option) (*Jitter, error) {
	if inner == nil {
		return nil, fmt.Errorf("jitter: inner provider must not be nil")
	}
	if max <= 0 {
		return nil, fmt.Errorf("jitter: max delay must be positive")
	}
	j := &Jitter{
		inner: inner,
		max:   max,
		sleep: func(ctx context.Context, d time.Duration) error {
			select {
			case <-time.After(d):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	for _, o := range opts {
		o(j)
	}
	return j, nil
}

func (j *Jitter) delay(ctx context.Context) error {
	//nolint:gosec // non-cryptographic random is intentional here
	d := time.Duration(rand.Int63n(int64(j.max)))
	return j.sleep(ctx, d)
}

// Get delegates directly without adding jitter.
func (j *Jitter) Get(ctx context.Context, key string) (string, error) {
	return j.inner.Get(ctx, key)
}

// Put sleeps for a random duration then writes the secret.
func (j *Jitter) Put(ctx context.Context, key, value string) error {
	if err := j.delay(ctx); err != nil {
		return err
	}
	return j.inner.Put(ctx, key, value)
}

// Delete sleeps for a random duration then removes the secret.
func (j *Jitter) Delete(ctx context.Context, key string) error {
	if err := j.delay(ctx); err != nil {
		return err
	}
	return j.inner.Delete(ctx, key)
}

// List delegates directly without adding jitter.
func (j *Jitter) List(ctx context.Context) ([]string, error) {
	return j.inner.List(ctx)
}
