// Package metrics provides lightweight counters and gauges for
// tracking vaultshift operation outcomes (rotations, syncs, errors).
package metrics

import (
	"sync"
	"time"
)

// EventType classifies a recorded metric event.
type EventType string

const (
	EventRotation EventType = "rotation"
	EventSync      EventType = "sync"
	EventError     EventType = "error"
)

// Entry is a single recorded metric.
type Entry struct {
	Type      EventType
	Provider  string
	Key       string
	Success   bool
	Duration  time.Duration
	RecordedAt time.Time
}

// Collector accumulates metric entries in memory.
type Collector struct {
	mu      sync.RWMutex
	entries []Entry
}

// New returns an initialised Collector.
func New() *Collector {
	return &Collector{}
}

// Record appends a new metric entry. RecordedAt is set to now if zero.
func (c *Collector) Record(e Entry) {
	if e.RecordedAt.IsZero() {
		e.RecordedAt = time.Now().UTC()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
}

// All returns a snapshot of every recorded entry.
func (c *Collector) All() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Entry, len(c.entries))
	copy(out, c.entries)
	return out
}

// Summary returns aggregate counts grouped by EventType.
func (c *Collector) Summary() map[EventType]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	summary := make(map[EventType]int)
	for _, e := range c.entries {
		summary[e.Type]++
	}
	return summary
}

// Reset clears all recorded entries.
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = nil
}
