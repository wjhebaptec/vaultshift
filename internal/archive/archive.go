// Package archive provides versioned secret archiving for vaultshift.
// It stores point-in-time snapshots of secrets keyed by provider and secret
// name, retaining up to a configurable number of historical entries.
package archive

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Entry holds a single archived value for a secret.
type Entry struct {
	Value     string
	Provider  string
	Key       string
	ArchivedAt time.Time
}

// Archive stores historical secret values.
type Archive struct {
	mu      sync.RWMutex
	entries map[string][]Entry
	maxAge  int // maximum number of entries per key
}

// New creates an Archive that retains at most maxPerKey entries per secret.
// If maxPerKey is <= 0 it defaults to 10.
func New(maxPerKey int) *Archive {
	if maxPerKey <= 0 {
		maxPerKey = 10
	}
	return &Archive{
		entries: make(map[string][]Entry),
		maxAge:  maxPerKey,
	}
}

func storeKey(provider, key string) string {
	return fmt.Sprintf("%s::%s", provider, key)
}

// Store archives value for the given provider and key.
func (a *Archive) Store(provider, key, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sk := storeKey(provider, key)
	e := Entry{
		Value:      value,
		Provider:   provider,
		Key:        key,
		ArchivedAt: time.Now().UTC(),
	}
	a.entries[sk] = append(a.entries[sk], e)
	if len(a.entries[sk]) > a.maxAge {
		a.entries[sk] = a.entries[sk][len(a.entries[sk])-a.maxAge:]
	}
}

// List returns all archived entries for the given provider and key,
// ordered oldest-first.
func (a *Archive) List(provider, key string) []Entry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sk := storeKey(provider, key)
	out := make([]Entry, len(a.entries[sk]))
	copy(out, a.entries[sk])
	return out
}

// Latest returns the most recently archived entry for the given provider and key.
func (a *Archive) Latest(provider, key string) (Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sk := storeKey(provider, key)
	if len(a.entries[sk]) == 0 {
		return Entry{}, errors.New("archive: no entries found")
	}
	return a.entries[sk][len(a.entries[sk])-1], nil
}

// Purge removes all archived entries for the given provider and key.
func (a *Archive) Purge(provider, key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.entries, storeKey(provider, key))
}
