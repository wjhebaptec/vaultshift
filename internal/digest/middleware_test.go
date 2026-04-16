package digest_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/digest"
	"github.com/vaultshift/internal/provider/mock"
)

func newWrapped(t *testing.T) (*digest.VerifyingProvider, *mock.Provider) {
	t.Helper()
	m := mock.New()
	s, _ := digest.New([]byte("test-secret"))
	vp, err := digest.Wrap(m, s)
	if err != nil {
		t.Fatalf("Wrap error: %v", err)
	}
	return vp, m
}

func TestWrap_NilProvider_ReturnsError(t *testing.T) {
	s, _ := digest.New([]byte("x"))
	_, err := digest.Wrap(nil, s)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrap_NilSigner_ReturnsError(t *testing.T) {
	_, err := digest.Wrap(mock.New(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrapPut_AndGet_Succeeds(t *testing.T) {
	vp, _ := newWrapped(t)
	ctx := context.Background()
	if err := vp.Put(ctx, "key", "value"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := vp.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "value" {
		t.Fatalf("expected 'value', got %q", got)
	}
}

func TestWrapGet_TamperedValue_ReturnsError(t *testing.T) {
	vp, inner := newWrapped(t)
	ctx := context.Background()
	_ = vp.Put(ctx, "key", "original")
	// Tamper directly via inner provider
	_ = inner.Put(ctx, "key", "tampered")
	_, err := vp.Get(ctx, "key")
	if err == nil {
		t.Fatal("expected tamper detection error")
	}
}

func TestWrapDelete_RemovesDigest(t *testing.T) {
	vp, _ := newWrapped(t)
	ctx := context.Background()
	_ = vp.Put(ctx, "key", "value")
	if err := vp.Delete(ctx, "key"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestWrapList_DelegatesToInner(t *testing.T) {
	vp, _ := newWrapped(t)
	ctx := context.Background()
	_ = vp.Put(ctx, "a", "1")
	_ = vp.Put(ctx, "b", "2")
	keys, err := vp.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}
