// Package epoch tracks discrete rotation epochs for secrets,
// allowing consumers to detect when a secret has been rotated
// and which generation is currently active.
package epoch

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Entry holds metadata for a single epoch.
type Entry struct {
	Generation uint64    `json:"generation"`
	AdvancedAt time.Time `json:"advanced_at"`
	Note       string    `json:"note,omitempty"`
}

// Tracker maintains per-key epoch counters.
type Tracker struct {
	mu      sync.RWMutex
	epochs  map[string][]Entry
	maxKeep int
}

// New returns a Tracker that retains up to maxKeep historical entries per key.
// maxKeep must be >= 1.
func New(maxKeep int) (*Tracker, error) {
	if maxKeep < 1 {
		return nil, errors.New("epoch: maxKeep must be >= 1")
	}
	return &Tracker{
		epochs:  make(map[string][]Entry),
		maxKeep: maxKeep,
	}, nil
}

// Advance increments the epoch for key, recording an optional note.
// The first call for a key sets generation to 1.
func (t *Tracker) Advance(key, note string) (Entry, error) {
	if key == "" {
		return Entry{}, errors.New("epoch: key must not be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	var next uint64 = 1
	if hist := t.epochs[key]; len(hist) > 0 {
		next = hist[len(hist)-1].Generation + 1
	}
	e := Entry{
		Generation: next,
		AdvancedAt: time.Now().UTC(),
		Note:       note,
	}
	t.epochs[key] = append(t.epochs[key], e)
	if len(t.epochs[key]) > t.maxKeep {
		t.epochs[key] = t.epochs[key][len(t.epochs[key])-t.maxKeep:]
	}
	return e, nil
}

// Current returns the latest epoch entry for key.
func (t *Tracker) Current(key string) (Entry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	hist, ok := t.epochs[key]
	if !ok || len(hist) == 0 {
		return Entry{}, fmt.Errorf("epoch: no epoch recorded for key %q", key)
	}
	return hist[len(hist)-1], nil
}

// History returns all retained epoch entries for key, oldest first.
func (t *Tracker) History(key string) ([]Entry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	hist, ok := t.epochs[key]
	if !ok {
		return nil, fmt.Errorf("epoch: no epoch recorded for key %q", key)
	}
	out := make([]Entry, len(hist))
	copy(out, hist)
	return out, nil
}

// Reset removes all epoch history for key.
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.epochs, key)
}
