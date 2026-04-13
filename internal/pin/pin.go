// Package pin provides secret version pinning, allowing callers to lock
// a secret key to a specific version string and later verify or retrieve
// the pinned version.
package pin

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNotPinned is returned when no pin exists for the requested key.
var ErrNotPinned = errors.New("pin: key is not pinned")

// Entry holds the pinned version for a single secret key.
type Entry struct {
	Key       string
	Provider  string
	Version   string
	PinnedAt  time.Time
}

// Pinner stores and retrieves version pins.
type Pinner struct {
	mu   sync.RWMutex
	pins map[string]Entry
}

// New returns an initialised Pinner.
func New() *Pinner {
	return &Pinner{pins: make(map[string]Entry)}
}

func storeKey(provider, key string) string {
	return fmt.Sprintf("%s::%s", provider, key)
}

// Pin records a version pin for the given provider+key pair.
func (p *Pinner) Pin(provider, key, version string) error {
	if provider == "" {
		return errors.New("pin: provider must not be empty")
	}
	if key == "" {
		return errors.New("pin: key must not be empty")
	}
	if version == "" {
		return errors.New("pin: version must not be empty")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pins[storeKey(provider, key)] = Entry{
		Key:      key,
		Provider: provider,
		Version:  version,
		PinnedAt: time.Now().UTC(),
	}
	return nil
}

// Get returns the pinned Entry for the given provider+key, or ErrNotPinned.
func (p *Pinner) Get(provider, key string) (Entry, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	e, ok := p.pins[storeKey(provider, key)]
	if !ok {
		return Entry{}, ErrNotPinned
	}
	return e, nil
}

// Unpin removes the pin for the given provider+key.
func (p *Pinner) Unpin(provider, key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pins, storeKey(provider, key))
}

// IsPinned reports whether a pin exists for the given provider+key.
func (p *Pinner) IsPinned(provider, key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.pins[storeKey(provider, key)]
	return ok
}

// All returns a snapshot of every pinned entry.
func (p *Pinner) All() []Entry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]Entry, 0, len(p.pins))
	for _, e := range p.pins {
		out = append(out, e)
	}
	return out
}
