package sync

import (
	"context"
	"fmt"

	"vaultshift/internal/provider"
)

// Syncer handles syncing secrets across multiple providers
type Syncer struct {
	registry *provider.Registry
}

// New creates a new Syncer instance
func New(registry *provider.Registry) *Syncer {
	return &Syncer{
		registry: registry,
	}
}

// SyncSecret synchronizes a secret from source to multiple targets
func (s *Syncer) SyncSecret(ctx context.Context, secretKey, sourceProvider string, targetProviders []string) error {
	source, err := s.registry.GetProvider(sourceProvider)
	if err != nil {
		return fmt.Errorf("failed to get source provider %s: %w", sourceProvider, err)
	}

	value, err := source.GetSecret(ctx, secretKey)
	if err != nil {
		return fmt.Errorf("failed to get secret from source: %w", err)
	}

	for _, targetName := range targetProviders {
		target, err := s.registry.GetProvider(targetName)
		if err != nil {
			return fmt.Errorf("failed to get target provider %s: %w", targetName, err)
		}

		if err := target.PutSecret(ctx, secretKey, value); err != nil {
			return fmt.Errorf("failed to put secret to %s: %w", targetName, err)
		}
	}

	return nil
}

// SyncAll synchronizes all secrets from source to targets
func (s *Syncer) SyncAll(ctx context.Context, sourceProvider string, targetProviders []string) error {
	source, err := s.registry.GetProvider(sourceProvider)
	if err != nil {
		return fmt.Errorf("failed to get source provider %s: %w", sourceProvider, err)
	}

	keys, err := source.ListSecrets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list secrets from source: %w", err)
	}

	for _, key := range keys {
		if err := s.SyncSecret(ctx, key, sourceProvider, targetProviders); err != nil {
			return fmt.Errorf("failed to sync secret %s: %w", key, err)
		}
	}

	return nil
}

// SyncWithFilter synchronizes secrets matching a filter function
func (s *Syncer) SyncWithFilter(ctx context.Context, sourceProvider string, targetProviders []string, filter FilterFunc) error {
	source, err := s.registry.GetProvider(sourceProvider)
	if err != nil {
		return fmt.Errorf("failed to get source provider %s: %w", sourceProvider, err)
	}

	keys, err := source.ListSecrets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list secrets from source: %w", err)
	}

	for _, key := range keys {
		if filter(key) {
			if err := s.SyncSecret(ctx, key, sourceProvider, targetProviders); err != nil {
				return fmt.Errorf("failed to sync secret %s: %w", key, err)
			}
		}
	}

	return nil
}
