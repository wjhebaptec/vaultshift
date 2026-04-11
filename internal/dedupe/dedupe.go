// Package dedupe provides deduplication tracking for secret rotation and sync
// operations, preventing redundant writes when a value has not changed.
package dedupe

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// Store tracks content fingerprints per provider+key pair to detect duplicates.
type Store struct {
	mu      sync.RWMutex
	hashes  map[string]string
}

// New creates a new dedupe Store.
func New() *Store {
	return &Store{
		hashes: make(map[string]string),
	}
}

func storeKey(provider, key string) string {
	return fmt.Sprintf("%s::%s", provider, key)
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// IsDuplicate reports whether the given value for provider+key is identical to
// the last recorded value. It returns false (not a duplicate) if no previous
// record exists.
func (s *Store) IsDuplicate(provider, key, value string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prev, ok := s.hashes[storeKey(provider, key)]
	if !ok {
		return false
	}
	return prev == hash(value)
}

// Record stores the fingerprint of value for the given provider+key.
func (s *Store) Record(provider, key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashes[storeKey(provider, key)] = hash(value)
}

// Forget removes the stored fingerprint for provider+key.
func (s *Store) Forget(provider, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.hashes, storeKey(provider, key))
}

// Reset clears all stored fingerprints.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashes = make(map[string]string)
}

// Size returns the number of tracked entries.
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.hashes)
}
