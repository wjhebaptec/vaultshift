package elect

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// GuardedProvider wraps a provider.Provider and gates all mutating
// operations behind a leader check. Read operations are always allowed.
type GuardedProvider struct {
	inner     provider.Provider
	elector   *Elector
	candidate string
}

// Guard returns a GuardedProvider that only allows writes when candidate
// is the current leader according to e.
func Guard(p provider.Provider, e *Elector, candidate string) (*GuardedProvider, error) {
	if p == nil {
		return nil, fmt.Errorf("elect: provider must not be nil")
	}
	if e == nil {
		return nil, fmt.Errorf("elect: elector must not be nil")
	}
	if candidate == "" {
		return nil, fmt.Errorf("elect: candidate must not be empty")
	}
	return &GuardedProvider{inner: p, elector: e, candidate: candidate}, nil
}

func (g *GuardedProvider) requireLeader() error {
	leader, ok := g.elector.Leader()
	if !ok || leader != g.candidate {
		return fmt.Errorf("elect: %w — current leader: %q", ErrNotLeader, leader)
	}
	return nil
}

// Put writes a secret only if the candidate is the current leader.
func (g *GuardedProvider) Put(ctx context.Context, key, value string) error {
	if err := g.requireLeader(); err != nil {
		return err
	}
	return g.inner.Put(ctx, key, value)
}

// Get reads a secret (always allowed).
func (g *GuardedProvider) Get(ctx context.Context, key string) (string, error) {
	return g.inner.Get(ctx, key)
}

// Delete removes a secret only if the candidate is the current leader.
func (g *GuardedProvider) Delete(ctx context.Context, key string) error {
	if err := g.requireLeader(); err != nil {
		return err
	}
	return g.inner.Delete(ctx, key)
}

// List enumerates secrets (always allowed).
func (g *GuardedProvider) List(ctx context.Context) ([]string, error) {
	return g.inner.List(ctx)
}
