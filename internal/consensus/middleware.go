package consensus

import (
	"context"
	"errors"

	"github.com/vaultshift/internal/provider"
)

// GuardedProvider wraps a primary provider and enforces consensus reads.
// Write operations (Put, Delete) are delegated directly to the primary.
type GuardedProvider struct {
	primary  provider.Provider
	reader   *Reader
}

// NewGuardedProvider creates a provider that reads via consensus and writes to primary.
func NewGuardedProvider(primary provider.Provider, reader *Reader) (*GuardedProvider, error) {
	if primary == nil {
		return nil, errors.New("consensus: primary provider must not be nil")
	}
	if reader == nil {
		return nil, errors.New("consensus: reader must not be nil")
	}
	return &GuardedProvider{primary: primary, reader: reader}, nil
}

// GetSecret returns the secret value only when consensus is reached.
func (g *GuardedProvider) GetSecret(ctx context.Context, key string) (string, error) {
	return g.reader.Get(ctx, key)
}

// PutSecret writes the secret to the primary provider.
func (g *GuardedProvider) PutSecret(ctx context.Context, key, value string) error {
	return g.primary.PutSecret(ctx, key, value)
}

// DeleteSecret removes the secret from the primary provider.
func (g *GuardedProvider) DeleteSecret(ctx context.Context, key string) error {
	return g.primary.DeleteSecret(ctx, key)
}

// ListSecrets lists secrets from the primary provider.
func (g *GuardedProvider) ListSecrets(ctx context.Context) ([]string, error) {
	return g.primary.ListSecrets(ctx)
}
