package envelope

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// AESEncryptor implements Encryptor using AES-256-GCM.
type AESEncryptor struct {
	gcm cipher.AEAD
}

// NewAESEncryptor creates an AESEncryptor from a 32-byte key.
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New("envelope: AES key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("envelope: create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("envelope: create GCM: %w", err)
	}
	return &AESEncryptor{gcm: gcm}, nil
}

// Encrypt seals plaintext with AES-256-GCM, prepending the nonce.
func (a *AESEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("envelope: generate nonce: %w", err)
	}
	return a.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt opens an AES-256-GCM sealed message (nonce prepended).
func (a *AESEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	ns := a.gcm.NonceSize()
	if len(ciphertext) < ns {
		return nil, errors.New("envelope: ciphertext too short")
	}
	nonce, data := ciphertext[:ns], ciphertext[ns:]
	plaintext, err := a.gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("envelope: decrypt: %w", err)
	}
	return plaintext, nil
}
