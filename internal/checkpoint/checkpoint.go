// Package checkpoint provides resumable operation tracking for long-running
// secret rotation and sync jobs. It persists progress so that interrupted
// operations can continue from where they left off.
package checkpoint

import (
	"errors"
	"sync"
	"time"
)

// Status represents the current state of a checkpointed operation.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Entry records the progress of a single key within an operation.
type Entry struct {
	Key       string
	Status    Status
	Error     string
	UpdatedAt time.Time
}

// Checkpoint tracks per-operation progress across multiple keys.
type Checkpoint struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New creates a new empty Checkpoint.
func New() *Checkpoint {
	return &Checkpoint{
		entries: make(map[string]*Entry),
	}
}

// Mark records the status for a given key, setting UpdatedAt to now if zero.
func (c *Checkpoint) Mark(key string, status Status, errMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &Entry{
		Key:       key,
		Status:    status,
		Error:     errMsg,
		UpdatedAt: time.Now().UTC(),
	}
}

// Get returns the Entry for a key, or an error if not found.
func (c *Checkpoint) Get(key string) (*Entry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok {
		return nil, errors.New("checkpoint: key not found: " + key)
	}
	return e, nil
}

// Pending returns all keys whose status is StatusPending or StatusRunning.
func (c *Checkpoint) Pending() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for k, e := range c.entries {
		if e.Status == StatusPending || e.Status == StatusRunning {
			keys = append(keys, k)
		}
	}
	return keys
}

// Reset removes all entries from the checkpoint.
func (c *Checkpoint) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*Entry)
}

// Summary returns counts grouped by Status.
func (c *Checkpoint) Summary() map[Status]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[Status]int)
	for _, e := range c.entries {
		out[e.Status]++
	}
	return out
}
