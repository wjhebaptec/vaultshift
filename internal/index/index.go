// Package index provides a searchable in-memory index over secret keys,
// enabling fast lookups by prefix, suffix, or substring across providers.
package index

import (
	"fmt"
	"strings"
	"sync"
)

// Entry represents an indexed secret key belonging to a provider.
type Entry struct {
	Provider string
	Key      string
}

// Index maintains an in-memory index of secret keys per provider.
type Index struct {
	mu      sync.RWMutex
	entries []Entry
}

// New returns an empty Index.
func New() *Index {
	return &Index{}
}

// Add registers a key under the given provider name.
func (idx *Index) Add(provider, key string) error {
	if provider == "" {
		return fmt.Errorf("index: provider must not be empty")
	}
	if key == "" {
		return fmt.Errorf("index: key must not be empty")
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries = append(idx.entries, Entry{Provider: provider, Key: key})
	return nil
}

// Remove deletes all entries matching the given provider and key.
func (idx *Index) Remove(provider, key string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	kept := idx.entries[:0]
	for _, e := range idx.entries {
		if e.Provider != provider || e.Key != key {
			kept = append(kept, e)
		}
	}
	idx.entries = kept
}

// Search returns all entries whose key contains the given substring.
func (idx *Index) Search(substr string) []Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var results []Entry
	for _, e := range idx.entries {
		if strings.Contains(e.Key, substr) {
			results = append(results, e)
		}
	}
	return results
}

// SearchPrefix returns all entries whose key starts with the given prefix.
func (idx *Index) SearchPrefix(prefix string) []Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var results []Entry
	for _, e := range idx.entries {
		if strings.HasPrefix(e.Key, prefix) {
			results = append(results, e)
		}
	}
	return results
}

// All returns every indexed entry.
func (idx *Index) All() []Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]Entry, len(idx.entries))
	copy(out, idx.entries)
	return out
}

// Reset clears all entries from the index.
func (idx *Index) Reset() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries = nil
}
