package journal_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/journal"
)

func TestAppend_ValidEntry(t *testing.T) {
	j := journal.New()
	err := j.Append(journal.Entry{
		Kind:    journal.KindRotate,
		Key:     "db/password",
		Provider: "aws",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", j.Len())
	}
}

func TestAppend_MissingKind_ReturnsError(t *testing.T) {
	j := journal.New()
	err := j.Append(journal.Entry{Key: "db/password"})
	if err == nil {
		t.Fatal("expected error for missing kind")
	}
}

func TestAppend_MissingKey_ReturnsError(t *testing.T) {
	j := journal.New()
	err := j.Append(journal.Entry{Kind: journal.KindSync})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAppend_SetsTimestampIfZero(t *testing.T) {
	j := journal.New()
	before := time.Now().UTC()
	_ = j.Append(journal.Entry{Kind: journal.KindDelete, Key: "x"})
	all := j.All()
	if all[0].OccurredAt.Before(before) {
		t.Error("expected OccurredAt to be set to approximately now")
	}
}

func TestAppend_PreservesExistingTimestamp(t *testing.T) {
	j := journal.New()
	ts := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = j.Append(journal.Entry{Kind: journal.KindPromote, Key: "k", OccurredAt: ts})
	all := j.All()
	if !all[0].OccurredAt.Equal(ts) {
		t.Errorf("expected %v, got %v", ts, all[0].OccurredAt)
	}
}

func TestAppend_AssignsIDIfEmpty(t *testing.T) {
	j := journal.New()
	_ = j.Append(journal.Entry{Kind: journal.KindSync, Key: "s"})
	all := j.All()
	if all[0].ID == "" {
		t.Error("expected auto-assigned ID")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	j := journal.New()
	_ = j.Append(journal.Entry{Kind: journal.KindRotate, Key: "a"})
	_ = j.Append(journal.Entry{Kind: journal.KindSync, Key: "b"})
	all := j.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestFilter_ByKind(t *testing.T) {
	j := journal.New()
	_ = j.Append(journal.Entry{Kind: journal.KindRotate, Key: "a", Provider: "aws"})
	_ = j.Append(journal.Entry{Kind: journal.KindSync, Key: "b", Provider: "gcp"})
	_ = j.Append(journal.Entry{Kind: journal.KindRotate, Key: "c", Provider: "vault"})

	results := j.Filter(journal.KindRotate, "")
	if len(results) != 2 {
		t.Fatalf("expected 2 rotate entries, got %d", len(results))
	}
}

func TestFilter_ByProvider(t *testing.T) {
	j := journal.New()
	_ = j.Append(journal.Entry{Kind: journal.KindSync, Key: "a", Provider: "aws"})
	_ = j.Append(journal.Entry{Kind: journal.KindSync, Key: "b", Provider: "gcp"})

	results := j.Filter("", "aws")
	if len(results) != 1 {
		t.Fatalf("expected 1 aws entry, got %d", len(results))
	}
	if results[0].Provider != "aws" {
		t.Errorf("unexpected provider: %s", results[0].Provider)
	}
}

func TestFilter_ByKindAndProvider(t *testing.T) {
	j := journal.New()
	_ = j.Append(journal.Entry{Kind: journal.KindRotate, Key: "a", Provider: "aws"})
	_ = j.Append(journal.Entry{Kind: journal.KindRotate, Key: "b", Provider: "gcp"})
	_ = j.Append(journal.Entry{Kind: journal.KindSync, Key: "c", Provider: "aws"})

	results := j.Filter(journal.KindRotate, "aws")
	if len(results) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(results))
	}
}
