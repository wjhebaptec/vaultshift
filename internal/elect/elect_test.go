package elect

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestCampaign_BecomesLeader(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	if err := e.Campaign("node-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	leader, ok := e.Leader()
	if !ok || leader != "node-1" {
		t.Fatalf("expected node-1 to be leader, got %q ok=%v", leader, ok)
	}
}

func TestCampaign_BlocksSecondCandidate(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	_ = e.Campaign("node-1")
	if err := e.Campaign("node-2"); err != ErrLeaderExists {
		t.Fatalf("expected ErrLeaderExists, got %v", err)
	}
}

func TestCampaign_AlreadyLeader(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	_ = e.Campaign("node-1")
	if err := e.Campaign("node-1"); err != ErrAlreadyLeader {
		t.Fatalf("expected ErrAlreadyLeader, got %v", err)
	}
}

func TestCampaign_TakesOverAfterExpiry(t *testing.T) {
	now := time.Now()
	e, _ := New(WithTTL(5*time.Second), WithClock(fixedClock(now)))
	_ = e.Campaign("node-1")
	// advance clock past TTL
	e.clock = fixedClock(now.Add(10 * time.Second))
	if err := e.Campaign("node-2"); err != nil {
		t.Fatalf("expected takeover to succeed, got %v", err)
	}
	leader, _ := e.Leader()
	if leader != "node-2" {
		t.Fatalf("expected node-2, got %q", leader)
	}
}

func TestRenew_ExtendLease(t *testing.T) {
	e, _ := New(WithTTL(5 * time.Second))
	_ = e.Campaign("node-1")
	if err := e.Renew("node-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenew_NonLeaderFails(t *testing.T) {
	e, _ := New(WithTTL(5 * time.Second))
	_ = e.Campaign("node-1")
	if err := e.Renew("node-2"); err != ErrNotLeader {
		t.Fatalf("expected ErrNotLeader, got %v", err)
	}
}

func TestResign_ReleasesLease(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	_ = e.Campaign("node-1")
	if err := e.Resign("node-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := e.Leader()
	if ok {
		t.Fatal("expected no leader after resign")
	}
}

func TestResign_NonLeaderFails(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	_ = e.Campaign("node-1")
	if err := e.Resign("node-2"); err != ErrNotLeader {
		t.Fatalf("expected ErrNotLeader, got %v", err)
	}
}

func TestNew_InvalidTTL_ReturnsError(t *testing.T) {
	if _, err := New(WithTTL(-1 * time.Second)); err == nil {
		t.Fatal("expected error for non-positive TTL")
	}
}

func TestLeader_ExpiredLease_ReturnsFalse(t *testing.T) {
	now := time.Now()
	e, _ := New(WithTTL(5*time.Second), WithClock(fixedClock(now)))
	_ = e.Campaign("node-1")
	e.clock = fixedClock(now.Add(10 * time.Second))
	_, ok := e.Leader()
	if ok {
		t.Fatal("expected lease to be expired")
	}
}
