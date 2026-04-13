package hedge

import (
	"context"
	"time"

	"github.com/vaultshift/internal/provider"
)

// ReadHedger wraps a primary provider and issues hedged reads to a set of
// fallback providers. Writes always go to the primary only.
type ReadHedger struct {
	primary   provider.Provider
	fallbacks []provider.Provider
	delay     time.Duration
}

// WrapRead returns a ReadHedger that uses primary for writes and hedges reads
// across primary + fallbacks.
func WrapRead(primary provider.Provider, fallbacks []provider.Provider, delay time.Duration) *ReadHedger {
	return &ReadHedger{
		primary:   primary,
		fallbacks: fallbacks,
		delay:     delay,
	}
}

// Get issues a hedged read.
func (r *ReadHedger) Get(ctx context.Context, key string) (string, error) {
	all := append([]provider.Provider{r.primary}, r.fallbacks...)
	h, err := New(all, WithDelay(r.delay))
	if err != nil {
		return r.primary.Get(ctx, key)
	}
	return h.Get(ctx, key)
}

// Put delegates to the primary provider only.
func (r *ReadHedger) Put(ctx context.Context, key, value string) error {
	return r.primary.Put(ctx, key, value)
}

// Delete delegates to the primary provider only.
func (r *ReadHedger) Delete(ctx context.Context, key string) error {
	return r.primary.Delete(ctx, key)
}

// List delegates to the primary provider only.
func (r *ReadHedger) List(ctx context.Context) ([]string, error) {
	return r.primary.List(ctx)
}

// Close closes the primary provider.
func (r *ReadHedger) Close() error {
	return r.primary.Close()
}
