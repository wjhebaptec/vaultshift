package index_test

import (
	"testing"

	"github.com/vaultshift/internal/index"
)

func TestAdd_AndSearch(t *testing.T) {
	idx := index.New()
	if err := idx.Add("aws", "prod/db/password"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = idx.Add("gcp", "prod/api/key")
	_ = idx.Add("aws", "staging/db/password")

	results := idx.Search("db")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestAdd_EmptyProvider_ReturnsError(t *testing.T) {
	idx := index.New()
	if err := idx.Add("", "some/key"); err == nil {
		t.Fatal("expected error for empty provider")
	}
}

func TestAdd_EmptyKey_ReturnsError(t *testing.T) {
	idx := index.New()
	if err := idx.Add("aws", ""); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSearchPrefix_MatchesCorrectEntries(t *testing.T) {
	idx := index.New()
	_ = idx.Add("aws", "prod/db/password")
	_ = idx.Add("aws", "prod/api/key")
	_ = idx.Add("gcp", "staging/db/password")

	results := idx.SearchPrefix("prod/")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Provider != "aws" {
			t.Errorf("expected provider aws, got %s", r.Provider)
		}
	}
}

func TestRemove_DeletesMatchingEntry(t *testing.T) {
	idx := index.New()
	_ = idx.Add("aws", "prod/db/password")
	_ = idx.Add("aws", "prod/api/key")

	idx.Remove("aws", "prod/db/password")

	all := idx.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry after remove, got %d", len(all))
	}
	if all[0].Key != "prod/api/key" {
		t.Errorf("unexpected remaining key: %s", all[0].Key)
	}
}

func TestReset_ClearsAllEntries(t *testing.T) {
	idx := index.New()
	_ = idx.Add("aws", "prod/db/password")
	_ = idx.Add("gcp", "prod/api/key")

	idx.Reset()

	if len(idx.All()) != 0 {
		t.Fatal("expected empty index after reset")
	}
}

func TestSearch_NoMatch_ReturnsNil(t *testing.T) {
	idx := index.New()
	_ = idx.Add("aws", "prod/db/password")

	results := idx.Search("nonexistent")
	if results != nil {
		t.Fatalf("expected nil, got %v", results)
	}
}
