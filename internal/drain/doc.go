// Package drain provides graceful shutdown coordination for vaultshift
// operations.
//
// A Drainer tracks in-flight secret operations and ensures that all active
// work finishes before the process exits. Call Acquire before starting an
// operation and Release (typically via defer) when it completes. Call Drain
// to signal shutdown and block until all tracked operations have released or
// the configured timeout elapses.
//
// The drain.Wrap helper integrates the Drainer transparently into any
// provider.Provider, making it easy to add graceful-shutdown behaviour
// without modifying provider implementations.
//
// Example:
//
//	d := drain.New(drain.WithTimeout(15 * time.Second))
//	wrapped := drain.Wrap(myProvider, d)
//
//	// … use wrapped as a normal provider …
//
//	if err := d.Drain(ctx); err != nil {
//		log.Printf("drain warning: %v", err)
//	}
package drain
