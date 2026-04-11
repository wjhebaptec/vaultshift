// Package quota implements per-key rate limiting for secret operations
// within vaultshift.
//
// A Limiter tracks how many times a given key (e.g. a secret name or
// provider identifier) has been accessed within a rolling time window.
// Once the configured limit is reached, further calls to Allow return
// ErrQuotaExceeded until the window resets.
//
// Example usage:
//
//	limiter := quota.New(
//		quota.WithLimit(100),
//		quota.WithWindow(time.Minute),
//	)
//
//	if err := limiter.Allow(secretName); err != nil {
//		// back off or surface the error to the caller
//	}
package quota
