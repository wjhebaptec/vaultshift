package version

import (
	"testing"
)

func TestHistory_PushAndLatest(t *testing.T) {
	h := NewHistory(5)
	h.Push("v1", "initial")
	h.Push("v2", "updated")

	e, err := h.Latest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Value != "v2" {
		t.Errorf("expected v2, got %q", e.Value)
	}
	if e.Label != "updated" {
		t.Errorf("expected label 'updated', got %q", e.Label)
	}
}

func TestHistory_Previous(t *testing.T) {
	h := NewHistory(5)
	h.Push("v1", "first")
	h.Push("v2", "second")

	e, err := h.Previous()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Value != "v1" {
		t.Errorf("expected v1, got %q", e.Value)
	}
}

func TestHistory_PreviousNotAvailable(t *testing.T) {
	h := NewHistory(5)
	h.Push("v1", "only")

	_, err := h.Previous()
	if err == nil {
		t.Error("expected error when no previous version exists")
	}
}

func TestHistory_LatestEmpty(t *testing.T) {
	h := NewHistory(5)
	_, err := h.Latest()
	if err == nil {
		t.Error("expected error on empty history")
	}
}

func TestHistory_MaxSizeEviction(t *testing.T) {
	h := NewHistory(3)
	for i := 0; i < 5; i++ {
		h.Push("val", "")
	}
	if h.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", h.Len())
	}
}

func TestHistory_All(t *testing.T) {
	h := NewHistory(10)
	h.Push("a", "")
	h.Push("b", "")
	h.Push("c", "")

	all := h.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].Value != "a" || all[2].Value != "c" {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestNewHistory_DefaultMaxSize(t *testing.T) {
	h := NewHistory(0)
	for i := 0; i < 15; i++ {
		h.Push("x", "")
	}
	if h.Len() != 10 {
		t.Errorf("expected default max 10, got %d", h.Len())
	}
}

func TestHistory_AllReturnsSnapshot(t *testing.T) {
	// Verify that modifying the slice returned by All does not affect the history.
	h := NewHistory(5)
	h.Push("a", "")
	h.Push("b", "")

	all := h.All()
	all[0].Value = "mutated"

	all2 := h.All()
	if all2[0].Value == "mutated" {
		t.Error("All() should return a snapshot; mutating it should not affect history")
	}
}
