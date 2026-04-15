package evict

import (
	"context"

	"github.com/vaultshift/internal/provider"
)

// Wrap returns a provider.Provider that transparently applies LRU eviction
// on top of the given provider using the specified capacity.
//
// This is a convenience constructor so callers that work with the
// provider.Provider interface do not need to reference the concrete Cache type.
func Wrap(p provider.Provider, capacity int) (provider.Provider, error) {
	return New(p, capacity)
}

// wrapped adapts Cache to the provider.Provider interface explicitly.
// Cache already satisfies the interface; this thin adapter makes the
// intent clear when used in middleware chains.
type wrapped struct{ *Cache }

func (w *wrapped) Put(ctx context.Context, key, value string) error {
	return w.Cache.Put(ctx, key, value)
}

func (w *wrapped) Get(ctx context.Context, key string) (string, error) {
	return w.Cache.Get(ctx, key)
}

func (w *wrapped) Delete(ctx context.Context, key string) error {
	return w.Cache.Delete(ctx, key)
}

func (w *wrapped) List(ctx context.Context) ([]string, error) {
	return w.Cache.List(ctx)
}
