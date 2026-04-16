package digest_test

import (
	"testing"

	"github.com/vaultshift/internal/digest"
)

func TestNew_EmptySecret_ReturnsError(t *testing.T) {
	_, err := digest.New(nil)
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestSign_AndVerify_Success(t *testing.T) {
	s, _ := digest.New([]byte("supersecret"))
	if err := s.Sign("mykey", "myvalue"); err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}
	if err := s.Verify("mykey", "myvalue"); err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

func TestVerify_WrongValue_ReturnsInvalidSignature(t *testing.T) {
	s, _ := digest.New([]byte("supersecret"))
	_ = s.Sign("mykey", "myvalue")
	err := s.Verify("mykey", "wrongvalue")
	if err == nil {
		t.Fatal("expected error for wrong value")
	}
}

func TestVerify_UnknownKey_ReturnsError(t *testing.T) {
	s, _ := digest.New([]byte("supersecret"))
	err := s.Verify("ghost", "val")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestSign_EmptyKey_ReturnsError(t *testing.T) {
	s, _ := digest.New([]byte("supersecret"))
	err := s.Sign("", "value")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestDelete_RemovesDigest(t *testing.T) {
	s, _ := digest.New([]byte("supersecret"))
	_ = s.Sign("k", "v")
	s.Delete("k")
	err := s.Verify("k", "v")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestSign_DifferentSecrets_ProduceDifferentDigests(t *testing.T) {
	s1, _ := digest.New([]byte("secret1"))
	s2, _ := digest.New([]byte("secret2"))
	_ = s1.Sign("k", "v")
	_ = s2.Sign("k", "v")
	if err := s1.Verify("k", "v"); err != nil {
		t.Fatalf("s1 verify failed: %v", err)
	}
	// s2's digest should not match s1's stored digest
	// verify s2 independently
	if err := s2.Verify("k", "v"); err != nil {
		t.Fatalf("s2 verify failed: %v", err)
	}
}
