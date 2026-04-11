// Package checkpoint provides resumable operation tracking for vaultshift
// rotation and sync pipelines.
//
// A Checkpoint records per-key progress as one of four statuses:
//
//	StatusPending   – queued but not yet started
//	StatusRunning   – currently being processed
//	StatusCompleted – finished successfully
//	StatusFailed    – finished with an error
//
// Usage:
//
//	cp := checkpoint.New()
//
//	// Mark a key as in-progress
//	cp.Mark("db/password", checkpoint.StatusRunning, "")
//
//	// On success
//	cp.Mark("db/password", checkpoint.StatusCompleted, "")
//
//	// On failure
//	cp.Mark("db/password", checkpoint.StatusFailed, err.Error())
//
//	// Resume: only re-process keys that didn't finish
//	for _, key := range cp.Pending() {
//		// retry key
//	}
//
//	// Inspect overall progress
//	summary := cp.Summary()
package checkpoint
