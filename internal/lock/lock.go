// Package lock provides distributed locking primitives to prevent
// concurrent secret rotation conflicts across multiple vaultshift instances.
package lock

import (
	"errors"
	"sync"
	"time"
)

// ErrAlreadyLocked is returned when a secret key is already locked.
var ErrAlreadyLocked = errors.New("lock: key is already locked")

// ErrNotLocked is returned when attempting to release a key that is not locked.
var ErrNotLocked = errors.New("lock: key is not locked")

// Entry holds metadata about an acquired lock.
type Entry struct {
	Owner     string
	AcquiredAt time.Time
	TTL       time.Duration
}

// IsExpired reports whether the lock TTL has elapsed.
func (e Entry) IsExpired() bool {
	if e.TTL <= 0 {
		return false
	}
	return time.Since(e.AcquiredAt) > e.TTL
}

// Manager tracks in-memory locks keyed by secret name.
type Manager struct {
	mu    sync.Mutex
	locks map[string]Entry
}

// New returns a new Manager.
func New() *Manager {
	return &Manager{locks: make(map[string]Entry)}
}

// Acquire attempts to lock the given key for owner with the specified TTL.
// A TTL of 0 means the lock never expires automatically.
func (m *Manager) Acquire(key, owner string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.locks[key]; ok && !e.IsExpired() {
		return ErrAlreadyLocked
	}

	m.locks[key] = Entry{
		Owner:      owner,
		AcquiredAt: time.Now(),
		TTL:        ttl,
	}
	return nil
}

// Release removes the lock for the given key.
func (m *Manager) Release(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.locks[key]; !ok {
		return ErrNotLocked
	}
	delete(m.locks, key)
	return nil
}

// IsLocked reports whether the key is currently locked (and not expired).
func (m *Manager) IsLocked(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.locks[key]
	return ok && !e.IsExpired()
}

// Get returns the lock entry for a key, and whether it exists and is valid.
func (m *Manager) Get(key string) (Entry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.locks[key]
	if !ok || e.IsExpired() {
		return Entry{}, false
	}
	return e, true
}
