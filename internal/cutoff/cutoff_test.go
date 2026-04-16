package cutoff_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/cutoff"
)

func fixedClock(t time.Time) cutoff.Clock { return func() time.Time { return t } }

func TestNew_InvalidWindow_ReturnsError(t *testing.T) {
	_, err := cutoff.New(0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestAllow_WithinWindow_Succeeds(t *testing.T) {
	now := time.Now()
	g, _ := cutoff.New(time.Minute, cutoff.WithClock(fixedClock(now)))
	g.Mark("key", now.Add(-30*time.Second))
	if err := g.Allow("key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_BeforeWindow_ReturnsError(t *testing.T) {
	now := time.Now()
	g, _ := cutoff.New(time.Minute, cutoff.WithClock(fixedClock(now)))
	g.Mark("key", now.Add(-2*time.Minute))
	if err := g.Allow("key"); err != cutoff.ErrBeforeCutoff {
		t.Fatalf("expected ErrBeforeCutoff, got %v", err)
	}
}

func TestAllow_UnknownKey_ReturnsError(t *testing.T) {
	g, _ := cutoff.New(time.Minute)
	if err := g.Allow("missing"); err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestForget_RemovesRecord(t *testing.T) {
	now := time.Now()
	g, _ := cutoff.New(time.Minute, cutoff.WithClock(fixedClock(now)))
	g.Mark("key", now)
	g.Forget("key")
	if err := g.Allow("key"); err == nil {
		t.Fatal("expected error after forget")
	}
}

func TestAllow_ExactThreshold_Succeeds(t *testing.T) {
	now := time.Now()
	g, _ := cutoff.New(time.Minute, cutoff.WithClock(fixedClock(now)))
	// exactly at the threshold boundary (not before)
	g.Mark("key", now.Add(-time.Minute))
	// at == threshold, not Before, so should pass
	if err := g.Allow("key"); err != nil {
		t.Fatalf("unexpected error at exact threshold: %v", err)
	}
}
