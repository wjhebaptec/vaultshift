// Package evict implements a least-recently-used (LRU) eviction layer for
// secret providers.
//
// # Overview
//
// When operating at scale, repeatedly fetching secrets from remote backends
// (AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault) adds latency and
// cost. The evict package wraps any provider.Provider with an in-process LRU
// cache of bounded size. When the cache is full the least-recently-used entry
// is silently dropped, keeping memory usage predictable.
//
// # Usage
//
//	// Wrap an existing provider with a 128-entry LRU cache.
//	cached, err := evict.New(vaultProvider, 128)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Use cached exactly like any other provider.
//	val, err := cached.Get(ctx, "my/secret")
//
// Alternatively, use evict.Wrap to obtain a provider.Provider interface
// directly, which is convenient when composing middleware chains.
package evict
