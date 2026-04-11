package quota_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/quota"
)

func TestAllow_WithinLimit(t *testing.T) {
	l := quota.New(quota.WithLimit(3), quota.WithWindow(time.Minute))
	for i := 0; i < 3; i++ {
		if err := l.Allow("key1"); err != nil {
			t.Fatalf("expected nil error on attempt %d, got %v", i+1, err)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l := quota.New(quota.WithLimit(2), quota.WithWindow(time.Minute))
	_ = l.Allow("key1")
	_ = l.Allow("key1")
	if err := l.Allow("key1"); err != quota.ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestAllow_SeparateKeys_IndependentLimits(t *testing.T) {
	l := quota.New(quota.WithLimit(1), quota.WithWindow(time.Minute))
	if err := l.Allow("a"); err != nil {
		t.Fatalf("unexpected error for key a: %v", err)
	}
	if err := l.Allow("b"); err != nil {
		t.Fatalf("unexpected error for key b: %v", err)
	}
	if err := l.Allow("a"); err != quota.ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded for key a, got %v", err)
	}
}

func TestAllow_WindowExpiry_ResetsCount(t *testing.T) {
	l := quota.New(quota.WithLimit(1), quota.WithWindow(50*time.Millisecond))
	if err := l.Allow("k"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := l.Allow("k"); err != quota.ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
	time.Sleep(60 * time.Millisecond)
	if err := l.Allow("k"); err != nil {
		t.Fatalf("expected nil after window reset, got %v", err)
	}
}

func TestUsage_ReturnsCurrentCount(t *testing.T) {
	l := quota.New(quota.WithLimit(10), quota.WithWindow(time.Minute))
	_ = l.Allow("x")
	_ = l.Allow("x")
	count, end := l.Usage("x")
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	if end.IsZero() {
		t.Fatal("expected non-zero window end")
	}
}

func TestUsage_UnknownKey_ReturnsZero(t *testing.T) {
	l := quota.New()
	count, end := l.Usage("missing")
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
	if !end.IsZero() {
		t.Fatal("expected zero time for unknown key")
	}
}

func TestReset_ClearsEntry(t *testing.T) {
	l := quota.New(quota.WithLimit(1), quota.WithWindow(time.Minute))
	_ = l.Allow("r")
	l.Reset("r")
	if err := l.Allow("r"); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}
