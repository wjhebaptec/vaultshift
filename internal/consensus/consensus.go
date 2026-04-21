// Package consensus provides a quorum-based read strategy that requires
// a minimum number of providers to agree on a secret value before returning it.
package consensus

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// ErrNoConsensus is returned when providers do not agree on a value.
var ErrNoConsensus = errors.New("consensus: providers disagree on secret value")

// ErrInsufficientProviders is returned when fewer providers respond than required.
var ErrInsufficientProviders = errors.New("consensus: insufficient providers responded")

// Reader reads a secret value only when a quorum of providers agree.
type Reader struct {
	registry *provider.Registry
	providers []string
	minAgree  int
}

// New creates a Reader requiring minAgree providers to return the same value.
func New(reg *provider.Registry, providerNames []string, minAgree int) (*Reader, error) {
	if reg == nil {
		return nil, errors.New("consensus: registry must not be nil")
	}
	if len(providerNames) == 0 {
		return nil, errors.New("consensus: at least one provider required")
	}
	if minAgree < 1 || minAgree > len(providerNames) {
		return nil, fmt.Errorf("consensus: minAgree must be between 1 and %d", len(providerNames))
	}
	return &Reader{
		registry:  reg,
		providers: providerNames,
		minAgree:  minAgree,
	}, nil
}

// Get retrieves a secret key and returns its value only if at least minAgree
// providers return the same value. Returns ErrNoConsensus on disagreement.
func (r *Reader) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("consensus: key must not be empty")
	}

	counts := make(map[string]int)
	responded := 0

	for _, name := range r.providers {
		p, err := r.registry.Get(name)
		if err != nil {
			continue
		}
		val, err := p.GetSecret(ctx, key)
		if err != nil {
			continue
		}
		counts[val]++
		responded++
	}

	if responded < r.minAgree {
		return "", ErrInsufficientProviders
	}

	for val, count := range counts {
		if count >= r.minAgree {
			return val, nil
		}
	}

	return "", ErrNoConsensus
}

// Providers returns the configured provider names.
func (r *Reader) Providers() []string {
	out := make([]string, len(r.providers))
	copy(out, r.providers)
	return out
}

// MinAgree returns the minimum agreement threshold.
func (r *Reader) MinAgree() int {
	return r.minAgree
}
