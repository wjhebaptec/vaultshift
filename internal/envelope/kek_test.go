package envelope_test

import (
	"testing"

	"github.com/vaultshift/internal/envelope"
)

func TestAESEncryptor_InvalidKeyLength(t *testing.T) {
	_, err := envelope.NewAESEncryptor([]byte("short"))
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestAESEncryptor_EncryptDecrypt(t *testing.T) {
	enc, err := envelope.NewAESEncryptor(newAESKey(t))
	if err != nil {
		t.Fatalf("NewAESEncryptor: %v", err)
	}

	plaintext := []byte("my-dek-bytes-here")
	ct, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	got, err := enc.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Errorf("got %q, want %q", got, plaintext)
	}
}

func TestAESEncryptor_DecryptTooShort(t *testing.T) {
	enc, _ := envelope.NewAESEncryptor(newAESKey(t))
	_, err := enc.Decrypt([]byte{0x01})
	if err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestAESEncryptor_DecryptCorrupted(t *testing.T) {
	enc, _ := envelope.NewAESEncryptor(newAESKey(t))
	ct, _ := enc.Encrypt([]byte("data"))
	ct[len(ct)-1] ^= 0xFF // flip last byte
	_, err := enc.Decrypt(ct)
	if err == nil {
		t.Fatal("expected error for corrupted ciphertext")
	}
}

func TestAESEncryptor_UniqueCiphertexts(t *testing.T) {
	enc, _ := envelope.NewAESEncryptor(newAESKey(t))
	a, _ := enc.Encrypt([]byte("same"))
	b, _ := enc.Encrypt([]byte("same"))
	if string(a) == string(b) {
		t.Error("expected different ciphertexts due to random nonce")
	}
}
