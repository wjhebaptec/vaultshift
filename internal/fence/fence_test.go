package fence_test

import (
	"errors"
	"testing"

	"github.com/vaultshift/internal/fence"
)

func TestCheck_FirstWrite_Succeeds(t *testing.T) {
	f := fence.New()
	if err := f.Check("key", 1); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheck_StrictlyIncreasing_Succeeds(t *testing.T) {
	f := fence.New()
	f.Check("key", 5)
	if err := f.Check("key", 6); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheck_SameSeq_ReturnsOutOfOrder(t *testing.T) {
	f := fence.New()
	f.Check("key", 3)
	err := f.Check("key", 3)
	if !errors.Is(err, fence.ErrOutOfOrder) {
		t.Fatalf("expected ErrOutOfOrder, got %v", err)
	}
}

func TestCheck_LowerSeq_ReturnsOutOfOrder(t *testing.T) {
	f := fence.New()
	f.Check("key", 10)
	err := f.Check("key", 9)
	if !errors.Is(err, fence.ErrOutOfOrder) {
		t.Fatalf("expected ErrOutOfOrder, got %v", err)
	}
}

func TestLatest_ReturnsLastAccepted(t *testing.T) {
	f := fence.New()
	f.Check("k", 7)
	v, err := f.Latest("k")
	if err != nil || v != 7 {
		t.Fatalf("expected 7/nil, got %d/%v", v, err)
	}
}

func TestLatest_UnknownKey_ReturnsError(t *testing.T) {
	f := fence.New()
	_, err := f.Latest("missing")
	if !errors.Is(err, fence.ErrUnknownKey) {
		t.Fatalf("expected ErrUnknownKey, got %v", err)
	}
}

func TestReset_AllowsReuseOfSeq(t *testing.T) {
	f := fence.New()
	f.Check("k", 5)
	f.Reset("k")
	if err := f.Check("k", 5); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestResetAll_ClearsEverything(t *testing.T) {
	f := fence.New()
	f.Check("a", 1)
	f.Check("b", 2)
	f.ResetAll()
	if _, err := f.Latest("a"); !errors.Is(err, fence.ErrUnknownKey) {
		t.Fatal("expected ErrUnknownKey after ResetAll")
	}
}

func TestCheck_IndependentKeys(t *testing.T) {
	f := fence.New()
	f.Check("x", 100)
	if err := f.Check("y", 1); err != nil {
		t.Fatalf("keys should be independent, got %v", err)
	}
}
