package accesslog

import (
	"testing"
	"time"
)

func TestRecord_AppendsEntry(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "aws", Key: "db/pass", Operation: OpGet, Success: true})
	entries := l.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Provider != "aws" {
		t.Errorf("expected provider aws, got %s", entries[0].Provider)
	}
}

func TestRecord_SetsTimestampIfZero(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "gcp", Key: "k", Operation: OpPut, Success: true})
	if l.Entries()[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestRecord_PreservesExistingTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New()
	l.Record(Entry{Timestamp: ts, Provider: "vault", Key: "k", Operation: OpDelete, Success: true})
	if !l.Entries()[0].Timestamp.Equal(ts) {
		t.Error("expected original timestamp to be preserved")
	}
}

func TestFilter_ByProvider(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "aws", Key: "a", Operation: OpGet, Success: true})
	l.Record(Entry{Provider: "gcp", Key: "b", Operation: OpGet, Success: true})
	result := l.Filter("aws", "")
	if len(result) != 1 || result[0].Provider != "aws" {
		t.Errorf("expected 1 aws entry, got %d", len(result))
	}
}

func TestFilter_ByOperation(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "aws", Key: "a", Operation: OpGet, Success: true})
	l.Record(Entry{Provider: "aws", Key: "b", Operation: OpPut, Success: true})
	result := l.Filter("", OpPut)
	if len(result) != 1 || result[0].Operation != OpPut {
		t.Errorf("expected 1 put entry, got %d", len(result))
	}
}

func TestFilter_Combined(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "aws", Key: "a", Operation: OpGet, Success: true})
	l.Record(Entry{Provider: "aws", Key: "b", Operation: OpPut, Success: true})
	l.Record(Entry{Provider: "gcp", Key: "c", Operation: OpGet, Success: true})
	result := l.Filter("aws", OpGet)
	if len(result) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result))
	}
}

func TestReset_ClearsEntries(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "aws", Key: "k", Operation: OpList, Success: true})
	l.Reset()
	if len(l.Entries()) != 0 {
		t.Error("expected entries to be empty after reset")
	}
}

func TestEntries_ReturnsCopy(t *testing.T) {
	l := New()
	l.Record(Entry{Provider: "aws", Key: "k", Operation: OpGet, Success: true})
	copy1 := l.Entries()
	copy1[0].Provider = "mutated"
	if l.Entries()[0].Provider == "mutated" {
		t.Error("Entries should return a copy, not a reference")
	}
}
