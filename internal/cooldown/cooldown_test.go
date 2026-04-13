package cooldown_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/cooldown"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_InvalidPeriod_ReturnsError(t *testing.T) {
	_, err := cooldown.New(0)
	if err == nil {
		t.Fatal("expected error for zero period")
	}
}

func TestAllow_FirstCall_Succeeds(t *testing.T) {
	c, _ := cooldown.New(time.Minute)
	if err := c.Allow("key1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_WithinCooldown_ReturnsError(t *testing.T) {
	now := time.Now()
	c, _ := cooldown.New(time.Minute, cooldown.WithClock(fixedClock(now)))
	_ = c.Allow("key1")
	if err := c.Allow("key1"); err == nil {
		t.Fatal("expected ErrCooldownActive")
	}
}

func TestAllow_AfterCooldown_Succeeds(t *testing.T) {
	base := time.Now()
	calls := 0
	clock := func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(2 * time.Minute)
	}
	c, _ := cooldown.New(time.Minute, cooldown.WithClock(clock))
	_ = c.Allow("key1")
	if err := c.Allow("key1"); err != nil {
		t.Fatalf("expected success after cooldown, got: %v", err)
	}
}

func TestAllow_SeparateKeys_AreIndependent(t *testing.T) {
	now := time.Now()
	c, _ := cooldown.New(time.Minute, cooldown.WithClock(fixedClock(now)))
	_ = c.Allow("key1")
	if err := c.Allow("key2"); err != nil {
		t.Fatalf("key2 should not be affected by key1 cooldown: %v", err)
	}
}

func TestRemaining_NoCooldown_ReturnsZero(t *testing.T) {
	c, _ := cooldown.New(time.Minute)
	if r := c.Remaining("ghost"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestRemaining_ActiveCooldown_ReturnsPositive(t *testing.T) {
	now := time.Now()
	c, _ := cooldown.New(time.Minute, cooldown.WithClock(fixedClock(now)))
	_ = c.Allow("key1")
	if r := c.Remaining("key1"); r <= 0 {
		t.Fatalf("expected positive remaining, got %v", r)
	}
}

func TestReset_ClearsCooldown(t *testing.T) {
	now := time.Now()
	c, _ := cooldown.New(time.Minute, cooldown.WithClock(fixedClock(now)))
	_ = c.Allow("key1")
	c.Reset("key1")
	if err := c.Allow("key1"); err != nil {
		t.Fatalf("expected success after reset, got: %v", err)
	}
}
