// Package version tracks secret version history and provides helpers for
// comparing and rolling back to previous secret states.
package version

import (
	"errors"
	"sync"
	"time"
)

// Entry represents a single historical version of a secret value.
type Entry struct {
	Value     string
	CreatedAt time.Time
	Label     string
}

// History stores ordered version entries for a single secret key.
type History struct {
	mu      sync.RWMutex
	entries []Entry
	maxSize int
}

// NewHistory creates a History that retains at most maxSize versions.
// If maxSize is <= 0 it defaults to 10.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 10
	}
	return &History{maxSize: maxSize}
}

// Push appends a new version entry, evicting the oldest if capacity is reached.
func (h *History) Push(value, label string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	e := Entry{Value: value, Label: label, CreatedAt: time.Now().UTC()}
	h.entries = append(h.entries, e)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
}

// Latest returns the most recent version entry.
func (h *History) Latest() (Entry, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.entries) == 0 {
		return Entry{}, errors.New("version: no entries in history")
	}
	return h.entries[len(h.entries)-1], nil
}

// Previous returns the entry immediately before the latest, enabling rollback.
func (h *History) Previous() (Entry, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.entries) < 2 {
		return Entry{}, errors.New("version: no previous version available")
	}
	return h.entries[len(h.entries)-2], nil
}

// All returns a copy of all stored entries, oldest first.
func (h *History) All() []Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Entry, len(h.entries))
	copy(out, h.entries)
	return out
}

// Len returns the number of stored versions.
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}
