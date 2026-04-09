// Package diff provides secret drift detection between two provider states.
//
// It compares a source map of key/value secrets against a destination map and
// classifies each key as added, removed, updated, or unchanged.
//
// Typical usage:
//
//	// Fetch secrets from two providers.
//	srcSecrets, _ := srcProvider.ListSecrets(ctx)
//	dstSecrets, _ := dstProvider.ListSecrets(ctx)
//
//	// Build maps of key -> value for comparison.
//	srcMap := secretsToMap(srcSecrets)
//	dstMap := secretsToMap(dstSecrets)
//
//	// Compute drift.
//	changes := diff.Compare(srcMap, dstMap)
//	if diff.HasDrift(changes) {
//		for _, c := range changes {
//			fmt.Println(c)
//		}
//	}
//
// The package is intentionally stateless and free of provider dependencies so
// it can be used in dry-run modes, CI pipelines, or audit reporting.
package diff
