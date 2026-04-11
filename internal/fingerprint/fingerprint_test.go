package fingerprint_test

import (
	"testing"

	"github.com/vaultshift/internal/fingerprint"
)

func TestHash_Deterministic(t *testing.T) {
	f := fingerprint.New()
	a := f.Hash("supersecret")
	b := f.Hash("supersecret")
	if a != b {
		t.Fatalf("expected identical hashes, got %s and %s", a, b)
	}
}

func TestHash_DifferentValues(t *testing.T) {
	f := fingerprint.New()
	if f.Hash("abc") == f.Hash("xyz") {
		t.Fatal("expected different hashes for different values")
	}
}

func TestHash_WithPrefix_IsolatesNamespace(t *testing.T) {
	no := fingerprint.New()
	with := fingerprint.New(fingerprint.WithPrefix("prod"))
	if no.Hash("secret") == with.Hash("secret") {
		t.Fatal("prefix should alter the hash")
	}
}

func TestHashMap_ReturnsAllKeys(t *testing.T) {
	f := fingerprint.New()
	secrets := map[string]string{"db_pass": "hunter2", "api_key": "abc123"}
	fp := f.HashMap(secrets)
	if len(fp) != len(secrets) {
		t.Fatalf("expected %d entries, got %d", len(secrets), len(fp))
	}
	for k := range secrets {
		if _, ok := fp[k]; !ok {
			t.Fatalf("missing key %s in fingerprint map", k)
		}
	}
}

func TestChanged_DetectsAddedKey(t *testing.T) {
	prev := map[string]string{"a": "h1"}
	next := map[string]string{"a": "h1", "b": "h2"}
	got := fingerprint.Changed(prev, next)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("expected [b], got %v", got)
	}
}

func TestChanged_DetectsRemovedKey(t *testing.T) {
	prev := map[string]string{"a": "h1", "b": "h2"}
	next := map[string]string{"a": "h1"}
	got := fingerprint.Changed(prev, next)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("expected [b], got %v", got)
	}
}

func TestChanged_DetectsUpdatedValue(t *testing.T) {
	prev := map[string]string{"a": "old"}
	next := map[string]string{"a": "new"}
	got := fingerprint.Changed(prev, next)
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("expected [a], got %v", got)
	}
}

func TestChanged_NoChanges(t *testing.T) {
	prev := map[string]string{"a": "h1"}
	next := map[string]string{"a": "h1"}
	if got := fingerprint.Changed(prev, next); len(got) != 0 {
		t.Fatalf("expected no changes, got %v", got)
	}
}

func TestSummarise_Deterministic(t *testing.T) {
	fp := map[string]string{"a": "h1", "b": "h2"}
	if fingerprint.Summarise(fp) != fingerprint.Summarise(fp) {
		t.Fatal("Summarise should be deterministic")
	}
}

func TestSummarise_DiffersOnChange(t *testing.T) {
	a := map[string]string{"x": "h1"}
	b := map[string]string{"x": "h2"}
	if fingerprint.Summarise(a) == fingerprint.Summarise(b) {
		t.Fatal("different fingerprint maps should produce different summaries")
	}
}
