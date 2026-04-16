// Package tee provides a provider wrapper that duplicates read results
// to a secondary provider for observation or warm-up purposes.
package tee

import (
	"context"
	"errors"

	"github.com/vaultshift/internal/provider"
)

// Tee wraps a primary provider and mirrors Get results to a secondary.
type Tee struct {
	primary   provider.Provider
	secondary provider.Provider
	writeOnly bool
}

// Option configures a Tee.
type Option func(*Tee)

// WithWriteOnly mirrors Put/Delete operations only, skipping Get mirroring.
func WithWriteOnly() Option {
	return func(t *Tee) { t.writeOnly = true }
}

// New creates a Tee that mirrors operations from primary to secondary.
func New(primary, secondary provider.Provider, opts ...Option) (*Tee, error) {
	if primary == nil {
		return nil, errors.New("tee: primary provider must not be nil")
	}
	if secondary == nil {
		return nil, errors.New("tee: secondary provider must not be nil")
	}
	t := &Tee{primary: primary, secondary: secondary}
	for _, o := range opts {
		o(t)
	}
	return t, nil
}

// Get retrieves from primary and mirrors the value to secondary unless write-only.
func (t *Tee) Get(ctx context.Context, key string) (string, error) {
	val, err := t.primary.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !t.writeOnly {
		_ = t.secondary.Put(ctx, key, val)
	}
	return val, nil
}

// Put writes to primary and mirrors to secondary.
func (t *Tee) Put(ctx context.Context, key, value string) error {
	if err := t.primary.Put(ctx, key, value); err != nil {
		return err
	}
	_ = t.secondary.Put(ctx, key, value)
	return nil
}

// Delete removes from primary and mirrors deletion to secondary.
func (t *Tee) Delete(ctx context.Context, key string) error {
	if err := t.primary.Delete(ctx, key); err != nil {
		return err
	}
	_ = t.secondary.Delete(ctx, key)
	return nil
}

// List delegates to primary only.
func (t *Tee) List(ctx context.Context) ([]string, error) {
	return t.primary.List(ctx)
}
