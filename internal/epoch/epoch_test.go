package epoch

import (
	"testing"
)

func TestNew_InvalidMaxKeep_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for maxKeep=0")
	}
}

func TestAdvance_FirstCall_GenerationOne(t *testing.T) {
	tr, _ := New(5)
	e, err := tr.Advance("mykey", "initial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Generation != 1 {
		t.Fatalf("expected generation 1, got %d", e.Generation)
	}
	if e.Note != "initial" {
		t.Fatalf("expected note 'initial', got %q", e.Note)
	}
	if e.AdvancedAt.IsZero() {
		t.Fatal("expected AdvancedAt to be set")
	}
}

func TestAdvance_IncrementsGeneration(t *testing.T) {
	tr, _ := New(5)
	tr.Advance("k", "")
	tr.Advance("k", "")
	e, _ := tr.Advance("k", "third")
	if e.Generation != 3 {
		t.Fatalf("expected generation 3, got %d", e.Generation)
	}
}

func TestAdvance_EmptyKey_ReturnsError(t *testing.T) {
	tr, _ := New(5)
	_, err := tr.Advance("", "note")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestCurrent_ReturnsLatest(t *testing.T) {
	tr, _ := New(5)
	tr.Advance("k", "first")
	tr.Advance("k", "second")
	e, err := tr.Current("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Generation != 2 {
		t.Fatalf("expected generation 2, got %d", e.Generation)
	}
}

func TestCurrent_UnknownKey_ReturnsError(t *testing.T) {
	tr, _ := New(5)
	_, err := tr.Current("missing")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestHistory_ReturnsCopiedSlice(t *testing.T) {
	tr, _ := New(5)
	tr.Advance("k", "a")
	tr.Advance("k", "b")
	hist, err := tr.History("k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(hist))
	}
}

func TestHistory_MaxKeep_Evicts(t *testing.T) {
	tr, _ := New(3)
	for i := 0; i < 5; i++ {
		tr.Advance("k", "")
	}
	hist, _ := tr.History("k")
	if len(hist) != 3 {
		t.Fatalf("expected 3 retained entries, got %d", len(hist))
	}
	if hist[0].Generation != 3 {
		t.Fatalf("expected oldest retained generation=3, got %d", hist[0].Generation)
	}
}

func TestReset_ClearsHistory(t *testing.T) {
	tr, _ := New(5)
	tr.Advance("k", "")
	tr.Reset("k")
	_, err := tr.Current("k")
	if err == nil {
		t.Fatal("expected error after reset")
	}
}
