package prefix

import (
	"context"

	"github.com/vaultshift/internal/provider"
)

// Wrapped is a provider.Provider that transparently prepends a namespace
// prefix to every key before delegating to the underlying provider.
type Wrapped struct {
	inner provider.Provider
	pfx   *Prefixer
}

// Wrap returns a new provider that qualifies every key with the given
// namespace using the supplied Prefixer.
func Wrap(p provider.Provider, pfx *Prefixer) (*Wrapped, error) {
	if p == nil {
		return nil, ErrEmptyNamespace
	}
	if pfx == nil {
		return nil, ErrEmptyNamespace
	}
	return &Wrapped{inner: p, pfx: pfx}, nil
}

// Put stores a secret under the prefixed key.
func (w *Wrapped) Put(ctx context.Context, key, value string) error {
	return w.inner.Put(ctx, w.pfx.Wrap(key), value)
}

// Get retrieves a secret by its un-prefixed key.
func (w *Wrapped) Get(ctx context.Context, key string) (string, error) {
	return w.inner.Get(ctx, w.pfx.Wrap(key))
}

// Delete removes a secret by its un-prefixed key.
func (w *Wrapped) Delete(ctx context.Context, key string) error {
	return w.inner.Delete(ctx, w.pfx.Wrap(key))
}

// List returns all keys stored under the namespace, with the prefix stripped.
func (w *Wrapped) List(ctx context.Context) ([]string, error) {
	all, err := w.inner.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(all))
	for _, k := range all {
		if stripped, ok := w.pfx.Unwrap(k); ok {
			out = append(out, stripped)
		}
	}
	return out, nil
}

// Close closes the underlying provider.
func (w *Wrapped) Close() error { return w.inner.Close() }
