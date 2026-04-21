// Package tenure tracks how long a secret has existed in a provider,
// enabling age-based policies and rotation scheduling.
package tenure

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrUnknownKey is returned when no tenure record exists for a key.
var ErrUnknownKey = errors.New("tenure: unknown key")

// Record holds the creation and last-seen timestamps for a secret.
type Record struct {
	Provider  string
	Key       string
	CreatedAt time.Time
	SeenAt    time.Time
}

// Age returns the duration since the secret was first registered.
func (r Record) Age() time.Duration {
	return time.Since(r.CreatedAt)
}

// Tracker records when secrets were first observed and last touched.
type Tracker struct {
	mu      sync.RWMutex
	records map[string]Record
	clock   func() time.Time
}

// New creates a new Tracker. Optionally supply a custom clock via WithClock.
func New(opts ...Option) *Tracker {
	t := &Tracker{
		records: make(map[string]Record),
		clock:   time.Now,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Option configures a Tracker.
type Option func(*Tracker)

// WithClock overrides the clock used for timestamping.
func WithClock(fn func() time.Time) Option {
	return func(t *Tracker) { t.clock = fn }
}

func storeKey(provider, key string) string {
	return fmt.Sprintf("%s\x00%s", provider, key)
}

// Touch registers or updates the seen-at timestamp for a secret.
// If the secret has not been seen before, CreatedAt is also set.
func (t *Tracker) Touch(provider, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.clock()
	k := storeKey(provider, key)
	if r, ok := t.records[k]; ok {
		r.SeenAt = now
		t.records[k] = r
	} else {
		t.records[k] = Record{
			Provider:  provider,
			Key:       key,
			CreatedAt: now,
			SeenAt:    now,
		}
	}
}

// Get returns the tenure record for the given provider+key pair.
func (t *Tracker) Get(provider, key string) (Record, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	r, ok := t.records[storeKey(provider, key)]
	if !ok {
		return Record{}, ErrUnknownKey
	}
	return r, nil
}

// Delete removes the tenure record for a secret.
func (t *Tracker) Delete(provider, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.records, storeKey(provider, key))
}

// OlderThan returns all records whose age exceeds the given threshold.
func (t *Tracker) OlderThan(d time.Duration) []Record {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Record
	for _, r := range t.records {
		if t.clock().Sub(r.CreatedAt) > d {
			out = append(out, r)
		}
	}
	return out
}
