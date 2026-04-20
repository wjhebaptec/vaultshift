// Package mirror provides a provider wrapper that replicates reads back to a
// secondary provider, keeping it in sync with the primary on every access.
package mirror

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Mirror wraps a primary provider and asynchronously (or synchronously)
// writes every value it reads back to a secondary provider.
type Mirror struct {
	primary   provider.Provider
	secondary provider.Provider
	writeBack bool // if true, Put also writes to secondary
}

// Option configures a Mirror.
type Option func(*Mirror)

// WithWriteBack enables mirroring on Put operations in addition to Get.
func WithWriteBack() Option {
	return func(m *Mirror) { m.writeBack = true }
}

// New creates a Mirror that reflects reads (and optionally writes) from
// primary to secondary.
func New(primary, secondary provider.Provider, opts ...Option) (*Mirror, error) {
	if primary == nil {
		return nil, errors.New("mirror: primary provider must not be nil")
	}
	if secondary == nil {
		return nil, errors.New("mirror: secondary provider must not be nil")
	}
	m := &Mirror{primary: primary, secondary: secondary}
	for _, o := range opts {
		o(m)
	}
	return m, nil
}

// Get retrieves a secret from the primary and mirrors the value to secondary.
func (m *Mirror) Get(ctx context.Context, key string) (string, error) {
	val, err := m.primary.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if putErr := m.secondary.Put(ctx, key, val); putErr != nil {
		return "", fmt.Errorf("mirror: failed to replicate key %q to secondary: %w", key, putErr)
	}
	return val, nil
}

// Put writes to the primary and, when write-back is enabled, also to secondary.
func (m *Mirror) Put(ctx context.Context, key, value string) error {
	if err := m.primary.Put(ctx, key, value); err != nil {
		return err
	}
	if m.writeBack {
		if err := m.secondary.Put(ctx, key, value); err != nil {
			return fmt.Errorf("mirror: write-back to secondary failed for key %q: %w", key, err)
		}
	}
	return nil
}

// Delete removes a key from both primary and secondary.
func (m *Mirror) Delete(ctx context.Context, key string) error {
	if err := m.primary.Delete(ctx, key); err != nil {
		return err
	}
	return m.secondary.Delete(ctx, key)
}

// List returns the keys from the primary provider.
func (m *Mirror) List(ctx context.Context) ([]string, error) {
	return m.primary.List(ctx)
}
