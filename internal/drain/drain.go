// Package drain provides graceful shutdown coordination for in-flight
// secret operations, ensuring all active work completes before teardown.
package drain

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Drainer tracks active operations and blocks shutdown until they complete
// or a deadline is exceeded.
type Drainer struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	closed  bool
	timeout time.Duration
}

// Option configures a Drainer.
type Option func(*Drainer)

// WithTimeout sets the maximum duration to wait for in-flight operations.
func WithTimeout(d time.Duration) Option {
	return func(dr *Drainer) {
		dr.timeout = d
	}
}

// New creates a new Drainer with optional configuration.
func New(opts ...Option) *Drainer {
	dr := &Drainer{
		timeout: 30 * time.Second,
	}
	for _, o := range opts {
		o(dr)
	}
	return dr
}

// Acquire registers a new in-flight operation. It returns an error if the
// Drainer has already been closed.
func (d *Drainer) Acquire() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return fmt.Errorf("drain: drainer is closed, cannot acquire")
	}
	d.wg.Add(1)
	return nil
}

// Release marks an in-flight operation as complete.
func (d *Drainer) Release() {
	d.wg.Done()
}

// Drain signals that no new operations should be accepted and waits for all
// in-flight operations to complete, respecting the configured timeout.
func (d *Drainer) Drain(ctx context.Context) error {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		return fmt.Errorf("drain: timed out waiting for in-flight operations: %w", timeoutCtx.Err())
	}
}

// IsClosed reports whether the Drainer has been closed.
func (d *Drainer) IsClosed() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.closed
}
