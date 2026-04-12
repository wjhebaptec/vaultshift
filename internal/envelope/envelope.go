// Package envelope provides secret envelope encryption, wrapping a data
// encryption key (DEK) with a key encryption key (KEK) so that secrets are
// never stored in plain-text alongside their encryption keys.
package envelope

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
)

// ErrNoKEK is returned when no key-encryption key is registered for a given ID.
var ErrNoKEK = errors.New("envelope: no KEK registered for id")

// Encryptor wraps and unwraps data encryption keys.
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// Sealed holds a DEK encrypted under a KEK together with the KEK id used.
type Sealed struct {
	KEKID        string `json:"kek_id"`
	EncryptedDEK string `json:"encrypted_dek"` // base64
	Ciphertext   string `json:"ciphertext"`    // base64, DEK-encrypted payload
}

// Manager manages multiple KEKs and performs envelope encryption.
type Manager struct {
	mu   sync.RWMutex
	keks map[string]Encryptor
}

// New returns a new Manager.
func New() *Manager {
	return &Manager{keks: make(map[string]Encryptor)}
}

// RegisterKEK registers an Encryptor under the given id.
func (m *Manager) RegisterKEK(id string, enc Encryptor) error {
	if id == "" {
		return errors.New("envelope: kek id must not be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keks[id] = enc
	return nil
}

// Seal encrypts plaintext using a freshly generated DEK, then wraps the DEK
// with the KEK identified by kekID.
func (m *Manager) Seal(kekID string, plaintext []byte) (*Sealed, error) {
	m.mu.RLock()
	kek, ok := m.keks[kekID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoKEK, kekID)
	}

	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("envelope: generate DEK: %w", err)
	}

	encryptedDEK, err := kek.Encrypt(dek)
	if err != nil {
		return nil, fmt.Errorf("envelope: wrap DEK: %w", err)
	}

	ciphertext, err := xorEncrypt(dek, plaintext)
	if err != nil {
		return nil, fmt.Errorf("envelope: encrypt payload: %w", err)
	}

	return &Sealed{
		KEKID:        kekID,
		EncryptedDEK: base64.StdEncoding.EncodeToString(encryptedDEK),
		Ciphertext:   base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// Open decrypts a Sealed envelope, recovering the original plaintext.
func (m *Manager) Open(s *Sealed) ([]byte, error) {
	m.mu.RLock()
	kek, ok := m.keks[s.KEKID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoKEK, s.KEKID)
	}

	encryptedDEK, err := base64.StdEncoding.DecodeString(s.EncryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("envelope: decode encrypted DEK: %w", err)
	}
	dek, err := kek.Decrypt(encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("envelope: unwrap DEK: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(s.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("envelope: decode ciphertext: %w", err)
	}
	return xorEncrypt(dek, ciphertext)
}

// xorEncrypt is a simple XOR stream cipher used to keep the package
// self-contained in tests; production callers should supply an AES-GCM KEK.
func xorEncrypt(key, data []byte) ([]byte, error) {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ key[i%len(key)]
	}
	return out, nil
}
