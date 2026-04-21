package scrub

import (
	"context"
	"errors"

	"github.com/vaultshift/internal/provider"
)

// wrapped is a provider.Provider that scrubs values on Put.
type wrapped struct {
	inner   provider.Provider
	scrubber *Scrubber
}

// Wrap returns a provider.Provider that scrubs secret values through s before
// delegating Put calls to inner. Get, Delete and List are passed through
// unchanged.
func Wrap(inner provider.Provider, s *Scrubber) (provider.Provider, error) {
	if inner == nil {
		return nil, errors.New("scrub: inner provider must not be nil")
	}
	if s == nil {
		return nil, errors.New("scrub: scrubber must not be nil")
	}
	return &wrapped{inner: inner, scrubber: s}, nil
}

func (w *wrapped) Put(ctx context.Context, key, value string) error {
	clean := w.scrubber.Clean(ctx, value)
	return w.inner.Put(ctx, key, clean)
}

func (w *wrapped) Get(ctx context.Context, key string) (string, error) {
	return w.inner.Get(ctx, key)
}

func (w *wrapped) Delete(ctx context.Context, key string) error {
	return w.inner.Delete(ctx, key)
}

func (w *wrapped) List(ctx context.Context) ([]string, error) {
	return w.inner.List(ctx)
}

func (w *wrapped) Close() error {
	return w.inner.Close()
}
