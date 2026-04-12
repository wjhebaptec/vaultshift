// Package journal provides an append-only operation journal for tracking
// secret lifecycle events with structured metadata.
package journal

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// EntryKind classifies the type of journal entry.
type EntryKind string

const (
	KindRotate   EntryKind = "rotate"
	KindSync     EntryKind = "sync"
	KindDelete   EntryKind = "delete"
	KindPromote  EntryKind = "promote"
	KindRollback EntryKind = "rollback"
)

// Entry represents a single recorded operation.
type Entry struct {
	ID         string
	Kind       EntryKind
	Provider   string
	Key        string
	Actor      string
	Message    string
	Meta       map[string]string
	OccurredAt time.Time
}

// Journal stores an ordered, append-only log of secret operation entries.
type Journal struct {
	mu      sync.RWMutex
	entries []Entry
	counter int
}

// New returns an empty Journal.
func New() *Journal {
	return &Journal{}
}

// Append adds a new entry to the journal. OccurredAt is set to now if zero.
func (j *Journal) Append(e Entry) error {
	if e.Kind == "" {
		return errors.New("journal: entry kind is required")
	}
	if e.Key == "" {
		return errors.New("journal: entry key is required")
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	j.counter++
	if e.ID == "" {
		e.ID = fmt.Sprintf("jrn-%d", j.counter)
	}
	j.entries = append(j.entries, e)
	return nil
}

// All returns a copy of all entries in insertion order.
func (j *Journal) All() []Entry {
	j.mu.RLock()
	defer j.mu.RUnlock()
	out := make([]Entry, len(j.entries))
	copy(out, j.entries)
	return out
}

// Filter returns entries matching the given kind and/or provider.
// Pass empty string to skip filtering on that field.
func (j *Journal) Filter(kind EntryKind, provider string) []Entry {
	j.mu.RLock()
	defer j.mu.RUnlock()
	var out []Entry
	for _, e := range j.entries {
		if kind != "" && e.Kind != kind {
			continue
		}
		if provider != "" && e.Provider != provider {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Len returns the total number of recorded entries.
func (j *Journal) Len() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return len(j.entries)
}
