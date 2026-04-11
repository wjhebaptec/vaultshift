// Package lineage provides secret provenance tracking for vaultshift.
//
// A Tracker records every operation performed on a secret — reads, rotations,
// syncs, exports — as an ordered sequence of Steps.  Each Step captures the
// provider involved, the operation name, the secret key, a timestamp, and
// optional metadata.
//
// Usage:
//
//	tr := lineage.New()
//
//	// Record that a secret was read from AWS.
//	tr.Add("db/password", lineage.Step{
//		Provider:  "aws",
//		Operation: "read",
//		Key:       "db/password",
//	})
//
//	// Later, record it being synced to Vault.
//	tr.Add("db/password", lineage.Step{
//		Provider:  "vault",
//		Operation: "sync",
//		Key:       "db/password",
//	})
//
//	// Retrieve the full history.
//	record, err := tr.Get("db/password")
//
// Records are safe for concurrent use.
package lineage
