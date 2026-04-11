// Package accesslog records provider access events for auditing and debugging.
package accesslog

import (
	"sync"
	"time"
)

// Operation represents the type of provider access.
type Operation string

const (
	OpGet    Operation = "get"
	OpPut    Operation = "put"
	OpDelete Operation = "delete"
	OpList   Operation = "list"
)

// Entry represents a single access log record.
type Entry struct {
	Timestamp  time.Time `json:"timestamp"`
	Provider   string    `json:"provider"`
	Key        string    `json:"key"`
	Operation  Operation `json:"operation"`
	Success    bool      `json:"success"`
	LatencyMs  int64     `json:"latency_ms"`
	Error      string    `json:"error,omitempty"`
}

// Log holds recorded access entries.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
}

// New creates a new empty access log.
func New() *Log {
	return &Log{}
}

// Record appends an access entry to the log.
// If Timestamp is zero it is set to the current UTC time.
func (l *Log) Record(e Entry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, e)
}

// Entries returns a copy of all recorded entries.
func (l *Log) Entries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Filter returns entries matching the given provider and/or operation.
// An empty string for provider or operation matches all values.
func (l *Log) Filter(provider string, op Operation) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if provider != "" && e.Provider != provider {
			continue
		}
		if op != "" && e.Operation != op {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Reset clears all recorded entries.
func (l *Log) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}
