// Package lock provides a lightweight in-memory distributed locking mechanism
// for vaultshift to coordinate secret rotation and sync operations.
//
// Locks are keyed by secret name and support optional TTL-based expiry,
// ensuring that stale locks from crashed workers are automatically released.
//
// Example usage:
//
//	m := lock.New()
//
//	if err := m.Acquire("db/password", "rotation-worker", 30*time.Second); err != nil {
//		// another worker holds the lock
//		return err
//	}
//	defer m.Release("db/password")
//
//	// perform rotation safely
package lock
