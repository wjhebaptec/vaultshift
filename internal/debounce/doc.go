// Package debounce coalesces rapid repeated triggers for the same secret key
// into a single deferred execution.
//
// # Overview
//
// When a secret watcher fires multiple change events in quick succession
// (e.g. during a rolling deploy), debounce ensures that the downstream
// handler is invoked only once — after the burst has settled.
//
// # Usage
//
//	d, err := debounce.New(500*time.Millisecond, func(ctx context.Context, key string) {
//		// called once per burst, with the key that changed
//		log.Printf("secret changed: %s", key)
//	})
//
//	// Each call resets the timer for that key.
//	d.Trigger(ctx, "db/password")
//	d.Trigger(ctx, "db/password") // resets
package debounce
