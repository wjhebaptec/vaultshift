// Package hedge provides a hedged request pattern for secret providers.
// A hedged request issues a secondary request after a delay if the primary
// has not yet responded, returning whichever result arrives first.
package hedge

import (
	"context"
	"errors"
	"time"

	"github.com/vaultshift/internal/provider"
)

// ErrNoProviders is returned when the hedger has no providers registered.
var ErrNoProviders = errors.New("hedge: no providers registered")

// Hedger issues reads to multiple providers with a configurable delay between
// each attempt and returns the first successful result.
type Hedger struct {
	providers []provider.Provider
	delay     time.Duration
}

// Option configures a Hedger.
type Option func(*Hedger)

// WithDelay sets the inter-hedge delay. Defaults to 50ms.
func WithDelay(d time.Duration) Option {
	return func(h *Hedger) {
		h.delay = d
	}
}

// New creates a Hedger backed by the given providers.
func New(providers []provider.Provider, opts ...Option) (*Hedger, error) {
	if len(providers) == 0 {
		return nil, ErrNoProviders
	}
	h := &Hedger{
		providers: providers,
		delay:     50 * time.Millisecond,
	}
	for _, o := range opts {
		o(h)
	}
	return h, nil
}

// Get returns the value for key from whichever provider responds first.
// Subsequent providers are queried after each delay interval.
func (h *Hedger) Get(ctx context.Context, key string) (string, error) {
	type result struct {
		val string
		err error
	}

	results := make(chan result, len(h.providers))

	for i, p := range h.providers {
		go func(idx int, prov provider.Provider) {
			if idx > 0 {
				select {
				case <-time.After(time.Duration(idx) * h.delay):
				case <-ctx.Done():
					results <- result{err: ctx.Err()}
					return
				}
			}
			v, err := prov.Get(ctx, key)
			results <- result{val: v, err: err}
		}(i, p)
	}

	var lastErr error
	for range h.providers {
		r := <-results
		if r.err == nil {
			return r.val, nil
		}
		lastErr = r.err
	}
	return "", lastErr
}
