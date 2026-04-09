// Package retry provides configurable retry logic with exponential backoff
// for operations that interact with remote secret manager providers.
//
// # Overview
//
// Use [New] with a [Config] to create a [Retryer]. Call [Retryer.Do] with
// a context and a function to execute. The function is retried up to
// MaxAttempts times, with an exponentially increasing delay between attempts
// capped at MaxDelay.
//
// # Example
//
//	cfg := retry.Config{
//		MaxAttempts:  4,
//		InitialDelay: 100 * time.Millisecond,
//		MaxDelay:     2 * time.Second,
//		Multiplier:   2.0,
//	}
//	r := retry.New(cfg)
//	err := r.Do(ctx, func() error {
//		return provider.PutSecret(key, value)
//	})
package retry
