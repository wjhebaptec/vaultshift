// Package expire provides TTL-based expiration tracking for secrets.
// It allows registering secrets with a time-to-live and querying
// which secrets have expired or are approaching expiration.
package expire

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds expiration metadata for a single secret.
type Entry struct {
	Key       string
	Provider  string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired reports whether the entry has passed its expiration time.
func (e Entry) IsExpired(now time.Time) bool {
	return now.After(e.ExpiresAt)
}

// ExpiresIn returns the duration remaining before expiration.
// Returns zero if already expired.
func (e Entry) ExpiresIn(now time.Time) time.Duration {
	d := e.ExpiresAt.Sub(now)
	if d < 0 {
		return 0
	}
	return d
}

// Tracker manages expiration entries for secrets.
type Tracker struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New creates a new Tracker.
func New() *Tracker {
	return &Tracker{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

func storeKey(provider, key string) string {
	return fmt.Sprintf("%s::%s", provider, key)
}

// Register records a secret with a given TTL.
func (t *Tracker) Register(provider, key string, ttl time.Duration) {
	now := t.now()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[storeKey(provider, key)] = Entry{
		Key:       key,
		Provider:  provider,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
}

// Get returns the expiration entry for a secret, if registered.
func (t *Tracker) Get(provider, key string) (Entry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.entries[storeKey(provider, key)]
	return e, ok
}

// Expired returns all entries that have passed their expiration time.
func (t *Tracker) Expired() []Entry {
	now := t.now()
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Entry
	for _, e := range t.entries {
		if e.IsExpired(now) {
			out = append(out, e)
		}
	}
	return out
}

// ExpiringSoon returns entries that will expire within the given window.
func (t *Tracker) ExpiringSoon(window time.Duration) []Entry {
	now := t.now()
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Entry
	for _, e := range t.entries {
		if !e.IsExpired(now) && e.ExpiresIn(now) <= window {
			out = append(out, e)
		}
	}
	return out
}

// Remove deletes a secret's expiration record.
func (t *Tracker) Remove(provider, key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, storeKey(provider, key))
}
