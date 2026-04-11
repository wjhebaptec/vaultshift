// Package encrypt provides symmetric encryption and decryption utilities
// for protecting secret values in transit or at rest within vaultshift.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// ErrInvalidKey is returned when the key length is not 16, 24, or 32 bytes.
var ErrInvalidKey = errors.New("encrypt: key must be 16, 24, or 32 bytes")

// ErrCiphertextTooShort is returned when the ciphertext is too short to contain a nonce.
var ErrCiphertextTooShort = errors.New("encrypt: ciphertext too short")

// Option configures an Encrypter.
type Option func(*Encrypter)

// Encrypter encrypts and decrypts values using AES-GCM.
type Encrypter struct {
	key []byte
}

// New creates a new Encrypter with the given key.
// The key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func New(key []byte, opts ...Option) (*Encrypter, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKey
	}
	e := &Encrypter{key: key}
	for _, o := range opts {
		o(e)
	}
	return e, nil
}

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext.
func (e *Encrypter) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts a base64-encoded ciphertext and returns the plaintext.
func (e *Encrypter) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", ErrCiphertextTooShort
	}
	nonce, cipherBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plainBytes, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}
