// Package watch implements continuous secret drift detection for vaultshift.
//
// A Watcher captures a baseline snapshot of secret values from a registered
// provider and then polls those same keys at a configurable interval. When a
// value is added, removed, or updated relative to the baseline, the supplied
// AlertFunc is invoked with the key and the diff.Result describing the change.
//
// Basic usage:
//
//	w := watch.New(registry, 30*time.Second, func(key string, r diff.Result) {
//		log.Printf("drift detected: %s (%s)", key, r.Status)
//	})
//	_ = w.Snapshot(ctx, "aws", []string{"prod/db/password"})
//	_ = w.Start(ctx, "aws", []string{"prod/db/password"})
//
// The watcher runs in a background goroutine and stops when the provided
// context is cancelled.
package watch
