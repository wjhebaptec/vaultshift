package label

import (
	"testing"
)

func TestSet_AndGet(t *testing.T) {
	m := New()
	err := m.Set("aws", "db/password", Labels{"env": "prod", "team": "platform"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := m.Get("aws", "db/password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["env"] != "prod" || got["team"] != "platform" {
		t.Errorf("unexpected labels: %v", got)
	}
}

func TestGet_UnknownKey(t *testing.T) {
	m := New()
	got, err := m.Get("aws", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil labels, got %v", got)
	}
}

func TestSet_EmptySecretKey_ReturnsError(t *testing.T) {
	m := New()
	err := m.Set("aws", "", Labels{"env": "prod"})
	if err == nil {
		t.Fatal("expected error for empty secret key")
	}
}

func TestSet_EmptyLabelKey_ReturnsError(t *testing.T) {
	m := New()
	err := m.Set("aws", "key", Labels{"": "value"})
	if err == nil {
		t.Fatal("expected error for empty label key")
	}
}

func TestSet_MergesLabels(t *testing.T) {
	m := New()
	_ = m.Set("aws", "key", Labels{"a": "1"})
	_ = m.Set("aws", "key", Labels{"b": "2"})
	got, _ := m.Get("aws", "key")
	if got["a"] != "1" || got["b"] != "2" {
		t.Errorf("expected merged labels, got %v", got)
	}
}

func TestSet_OverwritesExistingLabel(t *testing.T) {
	m := New()
	_ = m.Set("aws", "key", Labels{"env": "staging"})
	_ = m.Set("aws", "key", Labels{"env": "prod"})
	got, _ := m.Get("aws", "key")
	if got["env"] != "prod" {
		t.Errorf("expected overwritten label, got %v", got["env"])
	}
}

func TestDelete_RemovesLabels(t *testing.T) {
	m := New()
	_ = m.Set("aws", "key", Labels{"env": "prod"})
	_ = m.Delete("aws", "key")
	got, _ := m.Get("aws", "key")
	if got != nil {
		t.Errorf("expected nil after delete, got %v", got)
	}
}

func TestMatch_AllLabelsPresent(t *testing.T) {
	m := New()
	_ = m.Set("gcp", "secret/token", Labels{"env": "prod", "tier": "backend"})
	if !m.Match("gcp", "secret/token", Labels{"env": "prod"}) {
		t.Error("expected match")
	}
}

func TestMatch_MissingLabel(t *testing.T) {
	m := New()
	_ = m.Set("gcp", "secret/token", Labels{"env": "prod"})
	if m.Match("gcp", "secret/token", Labels{"tier": "backend"}) {
		t.Error("expected no match")
	}
}

func TestMatch_UnknownKey_ReturnsFalse(t *testing.T) {
	m := New()
	if m.Match("aws", "missing", Labels{"env": "prod"}) {
		t.Error("expected false for unknown key")
	}
}

func TestGet_ReturnsDefensiveCopy(t *testing.T) {
	m := New()
	_ = m.Set("aws", "key", Labels{"env": "prod"})
	got, _ := m.Get("aws", "key")
	got["env"] = "mutated"
	got2, _ := m.Get("aws", "key")
	if got2["env"] != "prod" {
		t.Error("expected defensive copy, original was mutated")
	}
}
