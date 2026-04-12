package shadow_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/shadow"
)

func setup(t *testing.T, mode shadow.Mode) (*shadow.Shadow, *mock.Provider, *mock.Provider) {
	t.Helper()
	prim := mock.New()
	shad := mock.New()
	s, err := shadow.New(prim, shad, mode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return s, prim, shad
}

func TestNew_NilPrimary_ReturnsError(t *testing.T) {
	_, err := shadow.New(nil, mock.New(), shadow.ModeWriteOnly)
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNew_NilShadow_ReturnsError(t *testing.T) {
	_, err := shadow.New(mock.New(), nil, shadow.ModeWriteOnly)
	if err == nil {
		t.Fatal("expected error for nil shadow")
	}
}

func TestPut_WritesBothProviders(t *testing.T) {
	s, prim, shad := setup(t, shadow.ModeWriteOnly)
	ctx := context.Background()

	if err := s.Put(ctx, "key1", "val1"); err != nil {
		t.Fatalf("put failed: %v", err)
	}

	v, _ := prim.Get(ctx, "key1")
	if v != "val1" {
		t.Errorf("primary: got %q, want %q", v, "val1")
	}
	v, _ = shad.Get(ctx, "key1")
	if v != "val1" {
		t.Errorf("shadow: got %q, want %q", v, "val1")
	}
}

func TestGet_WriteOnly_ReadsFromPrimary(t *testing.T) {
	s, prim, _, := setup(t, shadow.ModeWriteOnly)
	ctx := context.Background()
	_ = prim.Put(ctx, "k", "primary-value")

	v, err := s.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if v != "primary-value" {
		t.Errorf("got %q, want %q", v, "primary-value")
	}
	if len(s.Mismatches()) != 0 {
		t.Error("expected no mismatches in write-only mode")
	}
}

func TestGet_Compare_RecordsMismatch(t *testing.T) {
	s, prim, shad := setup(t, shadow.ModeCompare)
	ctx := context.Background()
	_ = prim.Put(ctx, "k", "primary-value")
	_ = shad.Put(ctx, "k", "shadow-value")

	v, err := s.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if v != "primary-value" {
		t.Errorf("got %q, want primary-value", v)
	}
	mm := s.Mismatches()
	if len(mm) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(mm))
	}
	if mm[0].Key != "k" || mm[0].Primary != "primary-value" || mm[0].Shadow != "shadow-value" {
		t.Errorf("unexpected mismatch: %+v", mm[0])
	}
}

func TestGet_Compare_NoMismatch_WhenEqual(t *testing.T) {
	s, prim, shad := setup(t, shadow.ModeCompare)
	ctx := context.Background()
	_ = prim.Put(ctx, "k", "same")
	_ = shad.Put(ctx, "k", "same")

	_, _ = s.Get(ctx, "k")
	if len(s.Mismatches()) != 0 {
		t.Error("expected no mismatches when values are equal")
	}
}

func TestResetMismatches_ClearsLog(t *testing.T) {
	s, prim, shad := setup(t, shadow.ModeCompare)
	ctx := context.Background()
	_ = prim.Put(ctx, "k", "a")
	_ = shad.Put(ctx, "k", "b")
	_, _ = s.Get(ctx, "k")

	s.ResetMismatches()
	if len(s.Mismatches()) != 0 {
		t.Error("expected empty mismatches after reset")
	}
}
