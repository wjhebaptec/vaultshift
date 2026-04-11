package encrypt_test

import (
	"strings"
	"testing"

	"github.com/vaultshift/internal/encrypt"
)

var validKey32 = []byte("12345678901234567890123456789012")
var validKey16 = []byte("1234567890123456")

func TestNew_ValidKey(t *testing.T) {
	_, err := encrypt.New(validKey32)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNew_InvalidKey(t *testing.T) {
	_, err := encrypt.New([]byte("short"))
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	e, _ := encrypt.New(validKey32)
	plaintext := "super-secret-value"

	ciphertext, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if ciphertext == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}

	result, err := e.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if result != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, result)
	}
}

func TestEncrypt_ProducesBase64(t *testing.T) {
	e, _ := encrypt.New(validKey16)
	out, err := e.Encrypt("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.ContainsAny(out, " \t\n") {
		t.Fatal("ciphertext should be valid base64 without whitespace")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	e, _ := encrypt.New(validKey32)
	_, err := e.Decrypt("!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	e, _ := encrypt.New(validKey32)
	// valid base64 but too short to hold a nonce
	_, err := e.Decrypt("aGk=")
	if err == nil {
		t.Fatal("expected ErrCiphertextTooShort")
	}
}

func TestEncrypt_NonDeterministic(t *testing.T) {
	e, _ := encrypt.New(validKey32)
	plaintext := "same-value"
	c1, _ := e.Encrypt(plaintext)
	c2, _ := e.Encrypt(plaintext)
	if c1 == c2 {
		t.Fatal("expected different ciphertexts due to random nonce")
	}
}

func TestEncryptDecrypt_AES128(t *testing.T) {
	e, _ := encrypt.New(validKey16)
	plain := "aes-128-test"
	ct, err := e.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	pt, err := e.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if pt != plain {
		t.Fatalf("expected %q got %q", plain, pt)
	}
}
