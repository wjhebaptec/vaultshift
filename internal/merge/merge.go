// Package merge provides utilities for combining secrets from multiple
// providers into a single unified map, with configurable conflict resolution.
package merge

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Strategy determines how conflicting keys are resolved.
type Strategy int

const (
	// StrategyFirst keeps the value from the first provider that defines the key.
	StrategyFirst Strategy = iota
	// StrategyLast overwrites with the value from the last provider that defines the key.
	StrategyLast
	// StrategyError returns an error when two providers define the same key.
	StrategyError
)

// Merger combines secrets from multiple providers.
type Merger struct {
	registry *provider.Registry
	strategy Strategy
}

// Option configures a Merger.
type Option func(*Merger)

// WithStrategy sets the conflict resolution strategy.
func WithStrategy(s Strategy) Option {
	return func(m *Merger) { m.strategy = s }
}

// New creates a Merger backed by the given registry.
func New(reg *provider.Registry, opts ...Option) *Merger {
	m := &Merger{registry: reg, strategy: StrategyFirst}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Merge lists all secrets from each named provider and combines them into a
// single map according to the configured conflict resolution strategy.
func (m *Merger) Merge(ctx context.Context, providerNames ...string) (map[string]string, error) {
	result := make(map[string]string)

	for _, name := range providerNames {
		p, err := m.registry.Get(name)
		if err != nil {
			return nil, fmt.Errorf("merge: provider %q not found: %w", name, err)
		}

		keys, err := p.ListSecrets(ctx)
		if err != nil {
			return nil, fmt.Errorf("merge: list secrets from %q: %w", name, err)
		}

		for _, key := range keys {
			val, err := p.GetSecret(ctx, key)
			if err != nil {
				return nil, fmt.Errorf("merge: get secret %q from %q: %w", key, name, err)
			}

			if existing, conflict := result[key]; conflict {
				switch m.strategy {
				case StrategyFirst:
					// keep existing — do nothing
					_ = existing
				case StrategyLast:
					result[key] = val
				case StrategyError:
					return nil, fmt.Errorf("merge: conflict on key %q", key)
				}
			} else {
				result[key] = val
			}
		}
	}

	return result, nil
}
