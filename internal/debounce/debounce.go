// Package debounce provides a mechanism to suppress rapid repeated secret
// operations, ensuring that only the final call within a quiet period is
// executed. This is useful when watching for secret changes that may arrive
// in bursts.
package debounce

import (
	"context"
	"sync"
	"time"
)

// Func is the function signature executed after the debounce window elapses.
type Func func(ctx context.Context, key string)

// Debouncer delays execution of a function until a quiet period has passed.
type Debouncer struct {
	mu      sync.Mutex
	timers  map[string]*time.Timer
	window  time.Duration
	fn      Func
}

// New creates a Debouncer that waits for window duration of inactivity before
// calling fn. window must be greater than zero.
func New(window time.Duration, fn Func) (*Debouncer, error) {
	if window <= 0 {
		return nil, fmt.Errorf("debounce: window must be greater than zero")
	}
	if fn == nil {
		return nil, fmt.Errorf("debounce: fn must not be nil")
	}
	return &Debouncer{
		timers: make(map[string]*time.Timer),
		window: window,
		fn:     fn,
	}, nil
}

// Trigger schedules fn to be called for key after the debounce window. If
// Trigger is called again for the same key before the window elapses, the
// timer resets.
func (d *Debouncer) Trigger(ctx context.Context, key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Stop()
	}

	d.timers[key] = time.AfterFunc(d.window, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
		d.fn(ctx, key)
	})
}

// Cancel stops a pending debounce for key without executing fn.
func (d *Debouncer) Cancel(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.timers[key]; ok {
		t.Stop()
		delete(d.timers, key)
	}
}

// Pending returns the set of keys that have pending debounced calls.
func (d *Debouncer) Pending() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	keys := make([]string, 0, len(d.timers))
	for k := range d.timers {
		keys = append(keys, k)
	}
	return keys
}
