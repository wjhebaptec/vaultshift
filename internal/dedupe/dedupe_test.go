package dedupe_test

import (
	"testing"

	"github.com/vaultshift/internal/dedupe"
)

func TestIsDuplicate_NoRecord_ReturnsFalse(t *testing.T) {
	s := dedupe.New()
	if s.IsDuplicate("aws", "my/secret", "value") {
		t.Fatal("expected false for unrecorded key")
	}
}

func TestRecord_ThenIsDuplicate_ReturnsTrue(t *testing.T) {
	s := dedupe.New()
	s.Record("aws", "my/secret", "supersecret")
	if !s.IsDuplicate("aws", "my/secret", "supersecret") {
		t.Fatal("expected duplicate to be detected")
	}
}

func TestIsDuplicate_DifferentValue_ReturnsFalse(t *testing.T) {
	s := dedupe.New()
	s.Record("aws", "my/secret", "old-value")
	if s.IsDuplicate("aws", "my/secret", "new-value") {
		t.Fatal("expected false for changed value")
	}
}

func TestIsDuplicate_SeparateProviders_AreIndependent(t *testing.T) {
	s := dedupe.New()
	s.Record("aws", "key", "val")
	if s.IsDuplicate("gcp", "key", "val") {
		t.Fatal("different providers should be independent")
	}
}

func TestForget_RemovesEntry(t *testing.T) {
	s := dedupe.New()
	s.Record("vault", "db/pass", "secret")
	s.Forget("vault", "db/pass")
	if s.IsDuplicate("vault", "db/pass", "secret") {
		t.Fatal("expected entry to be forgotten")
	}
}

func TestReset_ClearsAllEntries(t *testing.T) {
	s := dedupe.New()
	s.Record("aws", "a", "1")
	s.Record("aws", "b", "2")
	s.Reset()
	if s.Size() != 0 {
		t.Fatalf("expected size 0 after reset, got %d", s.Size())
	}
}

func TestSize_ReflectsRecordCount(t *testing.T) {
	s := dedupe.New()
	if s.Size() != 0 {
		t.Fatalf("expected initial size 0, got %d", s.Size())
	}
	s.Record("aws", "x", "v1")
	s.Record("gcp", "y", "v2")
	if s.Size() != 2 {
		t.Fatalf("expected size 2, got %d", s.Size())
	}
}

func TestRecord_Overwrite_UpdatesFingerprint(t *testing.T) {
	s := dedupe.New()
	s.Record("aws", "key", "old")
	s.Record("aws", "key", "new")
	if s.IsDuplicate("aws", "key", "old") {
		t.Fatal("old value should no longer be a duplicate after overwrite")
	}
	if !s.IsDuplicate("aws", "key", "new") {
		t.Fatal("new value should be detected as duplicate")
	}
}
