package archive

import (
	"testing"
	"time"
)

func TestStore_AndList(t *testing.T) {
	a := New(5)
	a.Store("aws", "db/pass", "secret1")
	a.Store("aws", "db/pass", "secret2")

	entries := a.List("aws", "db/pass")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Value != "secret1" || entries[1].Value != "secret2" {
		t.Error("unexpected entry order or values")
	}
}

func TestStore_SetsArchivedAt(t *testing.T) {
	a := New(5)
	before := time.Now().UTC()
	a.Store("gcp", "api/key", "val")
	after := time.Now().UTC()

	e, err := a.Latest("gcp", "api/key")
	if err != nil {
		t.Fatal(err)
	}
	if e.ArchivedAt.Before(before) || e.ArchivedAt.After(after) {
		t.Errorf("ArchivedAt %v not in expected range", e.ArchivedAt)
	}
}

func TestLatest_ReturnsNewest(t *testing.T) {
	a := New(5)
	a.Store("vault", "token", "old")
	a.Store("vault", "token", "new")

	e, err := a.Latest("vault", "token")
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "new" {
		t.Errorf("expected 'new', got %q", e.Value)
	}
}

func TestLatest_EmptyReturnsError(t *testing.T) {
	a := New(5)
	_, err := a.Latest("aws", "missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestStore_MaxPerKey_Evicts(t *testing.T) {
	a := New(3)
	for i := 0; i < 5; i++ {
		a.Store("aws", "k", "v")
	}
	if n := len(a.List("aws", "k")); n != 3 {
		t.Errorf("expected 3 entries after eviction, got %d", n)
	}
}

func TestPurge_RemovesEntries(t *testing.T) {
	a := New(5)
	a.Store("aws", "key", "val")
	a.Purge("aws", "key")
	if n := len(a.List("aws", "key")); n != 0 {
		t.Errorf("expected 0 entries after purge, got %d", n)
	}
}

func TestList_SeparateProviders_AreIndependent(t *testing.T) {
	a := New(5)
	a.Store("aws", "key", "aws-val")
	a.Store("gcp", "key", "gcp-val")

	awsEntries := a.List("aws", "key")
	gcpEntries := a.List("gcp", "key")

	if len(awsEntries) != 1 || awsEntries[0].Value != "aws-val" {
		t.Error("unexpected aws entries")
	}
	if len(gcpEntries) != 1 || gcpEntries[0].Value != "gcp-val" {
		t.Error("unexpected gcp entries")
	}
}

func TestNew_DefaultMaxPerKey(t *testing.T) {
	a := New(0)
	for i := 0; i < 15; i++ {
		a.Store("aws", "k", "v")
	}
	if n := len(a.List("aws", "k")); n != 10 {
		t.Errorf("expected default max 10, got %d", n)
	}
}
