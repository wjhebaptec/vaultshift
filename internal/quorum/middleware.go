package quorum

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// GuardedProvider wraps a provider.Provider and gates Put operations behind a
// quorum check across a set of peer providers. Get, Delete, and List are
// delegated directly to the primary provider.
type GuardedProvider struct {
	primary provider.Provider
	quorum  *Quorum
}

// NewGuardedProvider returns a GuardedProvider that requires quorum before
// committing any Put to the primary provider.
func NewGuardedProvider(primary provider.Provider, q *Quorum) (*GuardedProvider, error) {
	if primary == nil {
		return nil, fmt.Errorf("quorum: primary provider must not be nil")
	}
	if q == nil {
		return nil, fmt.Errorf("quorum: quorum must not be nil")
	}
	return &GuardedProvider{primary: primary, quorum: q}, nil
}

// Put writes the secret to all quorum providers and only proceeds to the
// primary if quorum is achieved.
func (g *GuardedProvider) Put(ctx context.Context, key, value string) error {
	if _, err := g.quorum.Put(ctx, "", key, value); err != nil {
		return fmt.Errorf("quorum guard: %w", err)
	}
	return g.primary.Put(ctx, key, value)
}

// Get delegates directly to the primary provider.
func (g *GuardedProvider) Get(ctx context.Context, key string) (string, error) {
	return g.primary.Get(ctx, key)
}

// Delete delegates directly to the primary provider.
func (g *GuardedProvider) Delete(ctx context.Context, key string) error {
	return g.primary.Delete(ctx, key)
}

// List delegates directly to the primary provider.
func (g *GuardedProvider) List(ctx context.Context) ([]string, error) {
	return g.primary.List(ctx)
}

// Name returns the primary provider's name with a quorum prefix.
func (g *GuardedProvider) Name() string {
	return "quorum:" + g.primary.Name()
}

// Close closes the primary provider.
func (g *GuardedProvider) Close() error {
	return g.primary.Close()
}
