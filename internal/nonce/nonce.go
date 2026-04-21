// Package nonce provides single-use token tracking to prevent replay attacks
// on secret operations across providers.
package nonce

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// ErrAlreadyUsed is returned when a nonce has already been consumed.
var ErrAlreadyUsed = errors.New("nonce: already used")

// ErrExpired is returned when a nonce is presented after its TTL has elapsed.
var ErrExpired = errors.New("nonce: expired")

// ErrUnknown is returned when a nonce is not found in the store.
var ErrUnknown = errors.New("nonce: unknown")

type entry struct {
	usedAt    time.Time
	expiresAt time.Time
	used      bool
}

// Store tracks issued nonces and their consumption state.
type Store struct {
	mu      sync.Mutex
	entries map[string]*entry
	ttl     time.Duration
	clock   func() time.Time
}

// Option configures a Store.
type Option func(*Store)

// WithTTL sets the lifetime of issued nonces.
func WithTTL(d time.Duration) Option {
	return func(s *Store) { s.ttl = d }
}

// WithClock overrides the time source (for testing).
func WithClock(fn func() time.Time) Option {
	return func(s *Store) { s.clock = fn }
}

// New creates a new nonce Store.
func New(opts ...Option) *Store {
	s := &Store{
		entries: make(map[string]*entry),
		ttl:     5 * time.Minute,
		clock:   time.Now,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Issue generates and registers a new nonce, returning its token.
func (s *Store) Issue() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	now := s.clock()
	s.mu.Lock()
	s.entries[token] = &entry{expiresAt: now.Add(s.ttl)}
	s.mu.Unlock()
	return token, nil
}

// Consume marks a nonce as used. Returns an error if unknown, expired, or already used.
func (s *Store) Consume(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[token]
	if !ok {
		return ErrUnknown
	}
	if s.clock().After(e.expiresAt) {
		delete(s.entries, token)
		return ErrExpired
	}
	if e.used {
		return ErrAlreadyUsed
	}
	e.used = true
	e.usedAt = s.clock()
	return nil
}

// Purge removes all expired entries from the store.
func (s *Store) Purge() int {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for k, e := range s.entries {
		if now.After(e.expiresAt) {
			delete(s.entries, k)
			removed++
		}
	}
	return removed
}
