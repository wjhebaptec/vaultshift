// Package fanout broadcasts a secret write to multiple provider targets
// concurrently. Each target receives the same key/value pair and its result
// is returned independently, allowing callers to inspect partial failures
// without blocking on a single slow or unhealthy provider.
//
// Basic usage:
//
//	f, err := fanout.New(registry, []string{"aws", "gcp", "vault"})
//	if err != nil {
//		log.Fatal(err)
//	}
//	results := f.Put(ctx, "db/password", newValue)
//	if fanout.HasFailures(results) {
//		for _, r := range results {
//			if r.Err != nil {
//				log.Printf("provider %s failed: %v", r.Provider, r.Err)
//			}
//		}
//	}
package fanout
