package mirror

import (
	"context"
	"fmt"
)

// SyncResult holds the outcome of a full synchronisation pass.
type SyncResult struct {
	Copied []string
	Errors []error
}

// HasFailures reports whether any errors occurred during sync.
func (r *SyncResult) HasFailures() bool { return len(r.Errors) > 0 }

// Sync copies every key from the primary to the secondary provider.
// It continues past individual key errors and collects them in the result.
func (m *Mirror) Sync(ctx context.Context) (*SyncResult, error) {
	keys, err := m.primary.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("mirror sync: listing primary: %w", err)
	}

	res := &SyncResult{}
	for _, k := range keys {
		val, err := m.primary.Get(ctx, k)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("mirror sync: read %q: %w", k, err))
			continue
		}
		if err := m.secondary.Put(ctx, k, val); err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("mirror sync: write %q: %w", k, err))
			continue
		}
		res.Copied = append(res.Copied, k)
	}
	return res, nil
}
