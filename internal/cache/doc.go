// Package cache implements a lightweight, thread-safe, TTL-based in-memory
// cache for secret values retrieved from remote secret managers.
//
// # Purpose
//
// Fetching secrets from AWS Secrets Manager, GCP Secret Manager, or HashiCorp
// Vault incurs network latency and may be subject to rate limits. The cache
// layer reduces redundant remote reads by storing recently fetched values
// for a configurable duration.
//
// # Usage
//
//	c := cache.New(5 * time.Minute)
//
//	// Store a secret
//	c.Set("db/password", "s3cr3t")
//
//	// Retrieve a secret
//	if val, ok := c.Get("db/password"); ok {
//		fmt.Println(val)
//	}
//
//	// Invalidate after rotation
//	c.Delete("db/password")
//
// A TTL of zero disables expiry and entries persist until explicitly deleted
// or the cache is flushed.
package cache
