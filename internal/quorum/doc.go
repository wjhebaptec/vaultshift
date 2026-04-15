// Package quorum implements a write-quorum strategy for secret managers.
//
// When writing a secret to multiple providers, a quorum requires that at least
// a configurable minimum number of providers acknowledge the write before the
// operation is considered successful. This prevents split-brain scenarios where
// only a subset of providers receive an update.
//
// Example usage:
//
//	q, err := quorum.New(2, awsProvider, gcpProvider, vaultProvider)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	failures, err := q.Put(ctx, "", "db/password", newSecret)
//	if err != nil {
//		log.Printf("quorum not met: %v", err)
//	}
//	for _, f := range failures {
//		log.Printf("provider %s failed: %v", f.Provider, f.Err)
//	}
package quorum
