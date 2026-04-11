package drain

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// WrappedProvider wraps a provider.Provider so that each operation is
// tracked by a Drainer, enabling graceful shutdown.
type WrappedProvider struct {
	inner   provider.Provider
	drainer *Drainer
}

// Wrap returns a WrappedProvider that tracks operations via the given Drainer.
func Wrap(p provider.Provider, d *Drainer) *WrappedProvider {
	return &WrappedProvider{inner: p, drainer: d}
}

// GetSecret acquires a drain slot, delegates to the inner provider, then releases.
func (w *WrappedProvider) GetSecret(ctx context.Context, key string) (string, error) {
	if err := w.drainer.Acquire(); err != nil {
		return "", fmt.Errorf("drain middleware: %w", err)
	}
	defer w.drainer.Release()
	return w.inner.GetSecret(ctx, key)
}

// PutSecret acquires a drain slot, delegates to the inner provider, then releases.
func (w *WrappedProvider) PutSecret(ctx context.Context, key, value string) error {
	if err := w.drainer.Acquire(); err != nil {
		return fmt.Errorf("drain middleware: %w", err)
	}
	defer w.drainer.Release()
	return w.inner.PutSecret(ctx, key, value)
}

// DeleteSecret acquires a drain slot, delegates to the inner provider, then releases.
func (w *WrappedProvider) DeleteSecret(ctx context.Context, key string) error {
	if err := w.drainer.Acquire(); err != nil {
		return fmt.Errorf("drain middleware: %w", err)
	}
	defer w.drainer.Release()
	return w.inner.DeleteSecret(ctx, key)
}

// ListSecrets acquires a drain slot, delegates to the inner provider, then releases.
func (w *WrappedProvider) ListSecrets(ctx context.Context) ([]string, error) {
	if err := w.drainer.Acquire(); err != nil {
		return nil, fmt.Errorf("drain middleware: %w", err)
	}
	defer w.drainer.Release()
	return w.inner.ListSecrets(ctx)
}

// Name returns the inner provider's name.
func (w *WrappedProvider) Name() string { return w.inner.Name() }

// Close delegates to the inner provider.
func (w *WrappedProvider) Close() error { return w.inner.Close() }
