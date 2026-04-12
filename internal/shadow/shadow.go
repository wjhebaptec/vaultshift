// Package shadow provides shadow-write functionality for safely migrating
// secrets between providers by writing to both a primary and shadow target.
package shadow

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Mode controls how shadow reads are handled.
type Mode int

const (
	// ModeWriteOnly writes to shadow but always reads from primary.
	ModeWriteOnly Mode = iota
	// ModeCompare writes to shadow and logs divergence on reads.
	ModeCompare
)

// Mismatch records a divergence between primary and shadow values.
type Mismatch struct {
	Key     string
	Primary string
	Shadow  string
}

// Shadow wraps a primary and shadow provider.
type Shadow struct {
	primary   provider.Provider
	shadow    provider.Provider
	mode      Mode
	mismatches []Mismatch
}

// New creates a Shadow instance.
func New(primary, shadow provider.Provider, mode Mode) (*Shadow, error) {
	if primary == nil {
		return nil, fmt.Errorf("shadow: primary provider must not be nil")
	}
	if shadow == nil {
		return nil, fmt.Errorf("shadow: shadow provider must not be nil")
	}
	return &Shadow{primary: primary, shadow: shadow, mode: mode}, nil
}

// Put writes to both primary and shadow providers.
func (s *Shadow) Put(ctx context.Context, key, value string) error {
	if err := s.primary.Put(ctx, key, value); err != nil {
		return fmt.Errorf("shadow: primary put failed: %w", err)
	}
	// Best-effort shadow write; do not fail the operation.
	_ = s.shadow.Put(ctx, key, value)
	return nil
}

// Get reads from primary; in ModeCompare it also reads from shadow and records mismatches.
func (s *Shadow) Get(ctx context.Context, key string) (string, error) {
	primVal, err := s.primary.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if s.mode == ModeCompare {
		shadowVal, serr := s.shadow.Get(ctx, key)
		if serr == nil && shadowVal != primVal {
			s.mismatches = append(s.mismatches, Mismatch{
				Key:     key,
				Primary: primVal,
				Shadow:  shadowVal,
			})
		}
	}
	return primVal, nil
}

// Mismatches returns all recorded divergences.
func (s *Shadow) Mismatches() []Mismatch {
	out := make([]Mismatch, len(s.mismatches))
	copy(out, s.mismatches)
	return out
}

// ResetMismatches clears the mismatch log.
func (s *Shadow) ResetMismatches() {
	s.mismatches = nil
}
