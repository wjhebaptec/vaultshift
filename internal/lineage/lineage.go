// Package lineage tracks the origin and transformation history of a secret,
// recording which provider it came from, when it was created, and any
// intermediate steps it passed through before reaching its destination.
package lineage

import (
	"errors"
	"sync"
	"time"
)

// Step represents a single hop in a secret's journey.
type Step struct {
	Provider  string    `json:"provider"`
	Operation string    `json:"operation"` // e.g. "read", "rotate", "sync", "export"
	Key       string    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Record holds the full lineage of a single secret key.
type Record struct {
	Key   string `json:"key"`
	Steps []Step `json:"steps"`
}

// Tracker stores lineage records keyed by secret name.
type Tracker struct {
	mu      sync.RWMutex
	records map[string]*Record
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{records: make(map[string]*Record)}
}

// Add appends a Step to the lineage of the given key.
// If no record exists for the key one is created automatically.
func (t *Tracker) Add(key string, step Step) {
	if step.Timestamp.IsZero() {
		step.Timestamp = time.Now().UTC()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.records[key]; !ok {
		t.records[key] = &Record{Key: key}
	}
	t.records[key].Steps = append(t.records[key].Steps, step)
}

// Get returns the Record for key, or an error if none exists.
func (t *Tracker) Get(key string) (Record, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	r, ok := t.records[key]
	if !ok {
		return Record{}, errors.New("lineage: no record for key " + key)
	}
	copy := *r
	copy.Steps = append([]Step(nil), r.Steps...)
	return copy, nil
}

// Keys returns all tracked secret keys.
func (t *Tracker) Keys() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	keys := make([]string, 0, len(t.records))
	for k := range t.records {
		keys = append(keys, k)
	}
	return keys
}

// Clear removes all lineage data.
func (t *Tracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.records = make(map[string]*Record)
}
