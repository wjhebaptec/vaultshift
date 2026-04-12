package shadow

import (
	"context"
	"fmt"
)

// PromoteResult holds the outcome of a single key promotion.
type PromoteResult struct {
	Key string
	Err error
}

// Promote copies all keys currently in the shadow provider into the primary
// provider, making the shadow the new source of truth.
// It returns a slice of results (one per key) and a summary error if any failed.
func (s *Shadow) Promote(ctx context.Context) ([]PromoteResult, error) {
	keys, err := s.shadow.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("shadow promote: list shadow keys: %w", err)
	}

	var results []PromoteResult
	var failed int

	for _, key := range keys {
		val, gerr := s.shadow.Get(ctx, key)
		if gerr != nil {
			results = append(results, PromoteResult{Key: key, Err: gerr})
			failed++
			continue
		}
		perr := s.primary.Put(ctx, key, val)
		results = append(results, PromoteResult{Key: key, Err: perr})
		if perr != nil {
			failed++
		}
	}

	if failed > 0 {
		return results, fmt.Errorf("shadow promote: %d key(s) failed", failed)
	}
	return results, nil
}
