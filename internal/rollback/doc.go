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
package rollback
