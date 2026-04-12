package envelope_test

import (
	"testing"

	"github.com/vaultshift/internal/envelope"
)

func newAESKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

func TestSealOpen_RoundTrip(t *testing.T) {
	m := envelope.New()
	enc, err := envelope.NewAESEncryptor(newAESKey(t))
	if err != nil {
		t.Fatalf("NewAESEncryptor: %v", err)
	}
	_ = m.RegisterKEK("primary", enc)

	plaintext := []byte("super-secret-value")
	sealed, err := m.Seal("primary", plaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := m.Open(sealed)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Errorf("got %q, want %q", got, plaintext)
	}
}

func TestSeal_UnknownKEK(t *testing.T) {
	m := envelope.New()
	_, err := m.Seal("missing", []byte("data"))
	if err == nil {
		t.Fatal("expected error for unknown KEK")
	}
}

func TestOpen_UnknownKEK(t *testing.T) {
	m := envelope.New()
	sealed := &envelope.Sealed{KEKID: "ghost"}
	_, err := m.Open(sealed)
	if err == nil {
		t.Fatal("expected error for unknown KEK")
	}
}

func TestRegisterKEK_EmptyID(t *testing.T) {
	m := envelope.New()
	enc, _ := envelope.NewAESEncryptor(newAESKey(t))
	if err := m.RegisterKEK("", enc); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestSeal_DifferentCiphertextsEachCall(t *testing.T) {
	m := envelope.New()
	enc, _ := envelope.NewAESEncryptor(newAESKey(t))
	_ = m.RegisterKEK("k1", enc)

	s1, _ := m.Seal("k1", []byte("hello"))
	s2, _ := m.Seal("k1", []byte("hello"))
	if s1.Ciphertext == s2.Ciphertext {
		t.Error("expected unique ciphertexts per Seal call")
	}
}

func TestSeal_MultipleKEKs(t *testing.T) {
	m := envelope.New()
	for _, id := range []string{"kek-a", "kek-b"} {
		enc, err := envelope.NewAESEncryptor(newAESKey(t))
		if err != nil {
			t.Fatalf("NewAESEncryptor: %v", err)
		}
		_ = m.RegisterKEK(id, enc)
	}

	for _, id := range []string{"kek-a", "kek-b"} {
		sealed, err := m.Seal(id, []byte("payload"))
		if err != nil {
			t.Fatalf("Seal(%s): %v", id, err)
		}
		got, err := m.Open(sealed)
		if err != nil {
			t.Fatalf("Open(%s): %v", id, err)
		}
		if string(got) != "payload" {
			t.Errorf("id=%s: got %q", id, got)
		}
	}
}
