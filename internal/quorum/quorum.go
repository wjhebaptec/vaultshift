// Package quorum requires agreement from a minimum number of providers
// before a secret write is considered successful.
package quorum

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// ErrQuorumNotMet is returned when fewer than the required number of providers
// acknowledge a write.
var ErrQuorumNotMet = errors.New("quorum: minimum acknowledgements not met")

// Quorum writes a secret to multiple providers and succeeds only when at least
// the configured minimum number of providers confirm the write.
type Quorum struct {
	providers []provider.Provider
	minAcks   int
}

// Failure records a provider name and the error it returned.
type Failure struct {
	Provider string
	Err      error
}

// New creates a Quorum that requires at least minAcks successful writes.
func New(minAcks int, providers ...provider.Provider) (*Quorum, error) {
	if minAcks < 1 {
		return nil, errors.New("quorum: minAcks must be at least 1")
	}
	if len(providers) == 0 {
		return nil, errors.New("quorum: at least one provider is required")
	}
	if minAcks > len(providers) {
		return nil, fmt.Errorf("quorum: minAcks (%d) exceeds provider count (%d)", minAcks, len(providers))
	}
	return &Quorum{providers: providers, minAcks: minAcks}, nil
}

// Put writes key/value to all providers and returns an error if fewer than
// minAcks providers succeed.
func (q *Quorum) Put(ctx context.Context, providerName, key, value string) ([]Failure, error) {
	var acks int
	var failures []Failure

	for _, p := range q.providers {
		names := p.Name()
		if providerName != "" && names != providerName {
			continue
		}
		if err := p.Put(ctx, key, value); err != nil {
			failures = append(failures, Failure{Provider: names, Err: err})
			continue
		}
		acks++
	}

	if acks < q.minAcks {
		return failures, fmt.Errorf("%w: got %d, need %d", ErrQuorumNotMet, acks, q.minAcks)
	}
	return failures, nil
}

// MinAcks returns the minimum acknowledgement threshold.
func (q *Quorum) MinAcks() int { return q.minAcks }

// ProviderCount returns the total number of registered providers.
func (q *Quorum) ProviderCount() int { return len(q.providers) }
