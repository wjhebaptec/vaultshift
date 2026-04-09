// Package rollback provides functionality to restore secrets to a previous version.
package rollback

import (
	"context"
	"fmt"
	"time"

	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/version"
)

// Record holds the information needed to undo a secret change.
type Record struct {
	Key       string
	PrevValue string
	RolledAt  time.Time
	Target    string
}

// Rollbacker restores secrets to a prior state using a version history.
type Rollbacker struct {
	registry *provider.Registry
	history  *version.History
}

// New creates a Rollbacker backed by the given registry and history store.
func New(registry *provider.Registry, history *version.History) *Rollbacker {
	return &Rollbacker{
		registry: registry,
		history:  history,
	}
}

// Rollback restores the secret at key in target to its most recent previous value.
func (r *Rollbacker) Rollback(ctx context.Context, target, key string) (*Record, error) {
	p, err := r.registry.Get(target)
	if err != nil {
		return nil, fmt.Errorf("rollback: provider %q not found: %w", target, err)
	}

	entry, err := r.history.Previous(key)
	if err != nil {
		return nil, fmt.Errorf("rollback: no previous version for key %q: %w", key, err)
	}

	if err := p.Put(ctx, key, entry.Value); err != nil {
		return nil, fmt.Errorf("rollback: failed to restore key %q: %w", key, err)
	}

	return &Record{
		Key:       key,
		PrevValue: entry.Value,
		RolledAt:  time.Now().UTC(),
		Target:    target,
	}, nil
}

// RollbackAll rolls back every key tracked in history for the given target.
func (r *Rollbacker) RollbackAll(ctx context.Context, target string) ([]*Record, error) {
	keys := r.history.Keys()
	var records []*Record
	var errs []error

	for _, key := range keys {
		rec, err := r.Rollback(ctx, target, key)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		records = append(records, rec)
	}

	if len(errs) > 0 {
		return records, fmt.Errorf("rollback: %d error(s) during rollback-all: %v", len(errs), errs)
	}
	return records, nil
}
