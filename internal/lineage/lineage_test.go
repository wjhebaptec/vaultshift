package lineage

import (
	"testing"
	"time"
)

func TestAdd_CreatesRecord(t *testing.T) {
	tr := New()
	tr.Add("db/password", Step{Provider: "aws", Operation: "read", Key: "db/password"})
	r, err := tr.Get("db/password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(r.Steps))
	}
	if r.Steps[0].Provider != "aws" {
		t.Errorf("expected provider aws, got %s", r.Steps[0].Provider)
	}
}

func TestAdd_SetsTimestampIfZero(t *testing.T) {
	tr := New()
	tr.Add("key", Step{Provider: "gcp", Operation: "rotate", Key: "key"})
	r, _ := tr.Get("key")
	if r.Steps[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestAdd_PreservesExistingTimestamp(t *testing.T) {
	tr := New()
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tr.Add("key", Step{Provider: "vault", Operation: "sync", Key: "key", Timestamp: fixed})
	r, _ := tr.Get("key")
	if !r.Steps[0].Timestamp.Equal(fixed) {
		t.Errorf("expected %v, got %v", fixed, r.Steps[0].Timestamp)
	}
}

func TestAdd_MultipleSteps(t *testing.T) {
	tr := New()
	for _, op := range []string{"read", "rotate", "sync"} {
		tr.Add("secret", Step{Provider: "aws", Operation: op, Key: "secret"})
	}
	r, _ := tr.Get("secret")
	if len(r.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(r.Steps))
	}
	if r.Steps[2].Operation != "sync" {
		t.Errorf("expected sync, got %s", r.Steps[2].Operation)
	}
}

func TestGet_UnknownKey(t *testing.T) {
	tr := New()
	_, err := tr.Get("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestKeys_ReturnsAll(t *testing.T) {
	tr := New()
	tr.Add("a", Step{Provider: "aws", Operation: "read", Key: "a"})
	tr.Add("b", Step{Provider: "gcp", Operation: "read", Key: "b"})
	keys := tr.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestClear_RemovesAll(t *testing.T) {
	tr := New()
	tr.Add("x", Step{Provider: "vault", Operation: "read", Key: "x"})
	tr.Clear()
	if len(tr.Keys()) != 0 {
		t.Error("expected empty tracker after Clear")
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	tr := New()
	tr.Add("k", Step{Provider: "aws", Operation: "read", Key: "k"})
	r, _ := tr.Get("k")
	r.Steps = append(r.Steps, Step{Provider: "mutated"})
	r2, _ := tr.Get("k")
	if len(r2.Steps) != 1 {
		t.Error("Get should return a copy; original was mutated")
	}
}
