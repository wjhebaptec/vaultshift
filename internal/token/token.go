// Package token provides short-lived token generation and validation
// for temporary secret access delegation.
package token

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// ErrTokenExpired is returned when a token has passed its TTL.
var ErrTokenExpired = errors.New("token: expired")

// ErrTokenNotFound is returned when a token does not exist in the store.
var ErrTokenNotFound = errors.New("token: not found")

// ErrTokenRevoked is returned when a token has been explicitly revoked.
var ErrTokenRevoked = errors.New("token: revoked")

// Entry holds metadata for an issued token.
type Entry struct {
	Token     string
	Scope     string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Revoked   bool
}

// IsExpired reports whether the token has passed its expiry time.
func (e *Entry) IsExpired(now time.Time) bool {
	return now.After(e.ExpiresAt)
}

// Manager issues and validates short-lived tokens.
type Manager struct {
	mu     sync.RWMutex
	store  map[string]*Entry
	ttl    time.Duration
	nowFn  func() time.Time
}

// Option configures a Manager.
type Option func(*Manager)

// WithTTL sets the default token time-to-live.
func WithTTL(d time.Duration) Option {
	return func(m *Manager) { m.ttl = d }
}

// WithClock injects a custom clock (useful for testing).
func WithClock(fn func() time.Time) Option {
	return func(m *Manager) { m.nowFn = fn }
}

// New creates a Manager with the provided options.
func New(opts ...Option) *Manager {
	m := &Manager{
		store: make(map[string]*Entry),
		ttl:   15 * time.Minute,
		nowFn: time.Now,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Issue generates a new token for the given scope.
func (m *Manager) Issue(scope string) (*Entry, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	now := m.nowFn()
	e := &Entry{
		Token:     hex.EncodeToString(b),
		Scope:     scope,
		IssuedAt:  now,
		ExpiresAt: now.Add(m.ttl),
	}
	m.mu.Lock()
	m.store[e.Token] = e
	m.mu.Unlock()
	return e, nil
}

// Validate checks that the token exists, is not revoked, and has not expired.
func (m *Manager) Validate(token string) (*Entry, error) {
	m.mu.RLock()
	e, ok := m.store[token]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrTokenNotFound
	}
	if e.Revoked {
		return nil, ErrTokenRevoked
	}
	if e.IsExpired(m.nowFn()) {
		return nil, ErrTokenExpired
	}
	return e, nil
}

// Revoke marks a token as revoked so it can no longer be validated.
func (m *Manager) Revoke(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.store[token]
	if !ok {
		return ErrTokenNotFound
	}
	e.Revoked = true
	return nil
}
