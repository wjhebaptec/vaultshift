package metrics_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/metrics"
)

func TestRecord_AppendsEntry(t *testing.T) {
	c := metrics.New()
	c.Record(metrics.Entry{Type: metrics.EventRotation, Provider: "aws", Key: "db/pass", Success: true})
	all := c.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	if all[0].Provider != "aws" {
		t.Errorf("unexpected provider: %s", all[0].Provider)
	}
}

func TestRecord_SetsTimestamp(t *testing.T) {
	c := metrics.New()
	before := time.Now().UTC()
	c.Record(metrics.Entry{Type: metrics.EventSync})
	after := time.Now().UTC()
	all := c.All()
	ts := all[0].RecordedAt
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v out of expected range [%v, %v]", ts, before, after)
	}
}

func TestRecord_PreservesExistingTimestamp(t *testing.T) {
	c := metrics.New()
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	c.Record(metrics.Entry{Type: metrics.EventError, RecordedAt: fixed})
	if got := c.All()[0].RecordedAt; !got.Equal(fixed) {
		t.Errorf("expected %v, got %v", fixed, got)
	}
}

func TestSummary_CountsByType(t *testing.T) {
	c := metrics.New()
	c.Record(metrics.Entry{Type: metrics.EventRotation})
	c.Record(metrics.Entry{Type: metrics.EventRotation})
	c.Record(metrics.Entry{Type: metrics.EventSync})
	c.Record(metrics.Entry{Type: metrics.EventError})
	s := c.Summary()
	if s[metrics.EventRotation] != 2 {
		t.Errorf("expected 2 rotations, got %d", s[metrics.EventRotation])
	}
	if s[metrics.EventSync] != 1 {
		t.Errorf("expected 1 sync, got %d", s[metrics.EventSync])
	}
}

func TestReset_ClearsEntries(t *testing.T) {
	c := metrics.New()
	c.Record(metrics.Entry{Type: metrics.EventSync})
	c.Reset()
	if len(c.All()) != 0 {
		t.Error("expected empty collector after reset")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	c := metrics.New()
	c.Record(metrics.Entry{Type: metrics.EventRotation, Key: "original"})
	snap := c.All()
	snap[0].Key = "mutated"
	if c.All()[0].Key != "original" {
		t.Error("All() should return a copy, not a reference")
	}
}
