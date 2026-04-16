// Package digest provides HMAC-based signing and verification for secret values.
package digest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// ErrInvalidSignature is returned when a signature does not match.
var ErrInvalidSignature = errors.New("digest: invalid signature")

// ErrUnknownKey is returned when no digest exists for a given key.
var ErrUnknownKey = errors.New("digest: unknown key")

// Signer signs and verifies secret values using HMAC-SHA256.
type Signer struct {
	mu     sync.RWMutex
	store  map[string]string // key -> hex digest
	secret []byte
}

// New creates a new Signer with the provided HMAC secret.
func New(secret []byte) (*Signer, error) {
	if len(secret) == 0 {
		return nil, errors.New("digest: secret must not be empty")
	}
	return &Signer{
		store:  make(map[string]string),
		secret: secret,
	}, nil
}

// Sign computes and stores the HMAC digest for the given key and value.
func (s *Signer) Sign(key, value string) error {
	if key == "" {
		return errors.New("digest: key must not be empty")
	}
	h := s.compute(value)
	s.mu.Lock()
	s.store[key] = h
	s.mu.Unlock()
	return nil
}

// Verify checks whether the given value matches the stored digest for key.
func (s *Signer) Verify(key, value string) error {
	s.mu.RLock()
	stored, ok := s.store[key]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownKey, key)
	}
	expected := s.compute(value)
	if !hmac.Equal([]byte(stored), []byte(expected)) {
		return fmt.Errorf("%w: %s", ErrInvalidSignature, key)
	}
	return nil
}

// Delete removes the stored digest for key.
func (s *Signer) Delete(key string) {
	s.mu.Lock()
	delete(s.store, key)
	s.mu.Unlock()
}

func (s *Signer) compute(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
