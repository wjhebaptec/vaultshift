// Package rollback implements secret rollback functionality for vaultshift.
//
// It allows operators to restore a secret to its most recent previous value
// as recorded in the version history, either for a single key or for all
// tracked keys within a target provider.
//
// Basic usage:
//
//	rb := rollback.New(registry, history)
//
//	// Restore a single key
//	rec, err := rb.Rollback(ctx, "aws", "prod/db/password")
//
//	// Restore all tracked keys
//	records, err := rb.RollbackAll(ctx, "aws")
//
// Rollback depends on version.History to retrieve the previous snapshot and
// provider.Registry to write the restored value back to the target provider.
//
// Error handling:
//
// If no previous version exists for a given key, Rollback returns
// ErrNoPreviousVersion. If the target provider is not found in the registry,
// ErrProviderNotFound is returned. Callers should check for these sentinel
// errors to distinguish between missing history and configuration issues:
//
//	rec, err := rb.Rollback(ctx, "aws", "prod/db/password")
//	if errors.Is(err, rollback.ErrNoPreviousVersion) {
//		// no prior version recorded for this key
//	}
//	if errors.Is(err, rollback.ErrProviderNotFound) {
//		// provider "aws" is not registered
//	}
package rollback
