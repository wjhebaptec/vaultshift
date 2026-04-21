package bridge

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// SyncResult captures the outcome of a bidirectional sync operation.
type SyncResult struct {
	AtoB []Result
	BtoA []Result
}

// HasFailures returns true if either direction recorded errors.
func (s SyncResult) HasFailures() bool {
	return HasFailures(s.AtoB) || HasFailures(s.BtoA)
}

// Sync performs a bidirectional sync between providerA and providerB.
// Keys present in A are written to B, and keys present in B are written to A.
func Sync(ctx context.Context, reg *provider.Registry, nameA, nameB string) (SyncResult, error) {
	bAB, err := New(reg, nameA, nameB)
	if err != nil {
		return SyncResult{}, fmt.Errorf("bridge sync A→B: %w", err)
	}
	bBA, err := New(reg, nameB, nameA)
	if err != nil {
		return SyncResult{}, fmt.Errorf("bridge sync B→A: %w", err)
	}

	atob, err := bAB.Forward(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("bridge sync A→B forward: %w", err)
	}
	btoa, err := bBA.Forward(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("bridge sync B→A forward: %w", err)
	}
	return SyncResult{AtoB: atob, BtoA: btoa}, nil
}
