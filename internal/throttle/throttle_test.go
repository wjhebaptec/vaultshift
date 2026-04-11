package throttle_test

import (
	"context"
	"testing"
	"time"

	"github.com/vaultshift/internal/throttle"
)

func TestAllow_WithinLimit(t *testing.T) {
	th := throttle.New(throttle.WithRate(5))
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := th.Allow(ctx, "key1"); err != nil {
			t.Fatalf("expected allow on attempt %d, got: %v", i+1, err)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	th := throttle.New(throttle.WithRate(3))
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = th.Allow(ctx, "key1")
	}
	err := th.Allow(ctx, "key1")
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}
}

func TestAllow_SeparateKeys_AreIndependent(t *testing.T) {
	th := throttle.New(throttle.WithRate(2))
	ctx := context.Background()

	_ = th.Allow(ctx, "a")
	_ = th.Allow(ctx, "a")

	if err := th.Allow(ctx, "b"); err != nil {
		t.Fatalf("key 'b' should not be throttled: %v", err)
	}
}

func TestAllow_ContextCancelled(t *testing.T) {
	th := throttle.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := th.Allow(ctx, "key1"); err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestUsage_ReturnsCurrentCount(t *testing.T) {
	th := throttle.New(throttle.WithRate(10))
	ctx := context.Background()

	_ = th.Allow(ctx, "key1")
	_ = th.Allow(ctx, "key1")
	_ = th.Allow(ctx, "key1")

	if got := th.Usage("key1"); got != 3 {
		t.Fatalf("expected usage 3, got %d", got)
	}
}

func TestReset_ClearsCount(t *testing.T) {
	th := throttle.New(throttle.WithRate(2))
	ctx := context.Background()

	_ = th.Allow(ctx, "key1")
	_ = th.Allow(ctx, "key1")

	th.Reset("key1")
	if got := th.Usage("key1"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestAllow_WindowExpiry_ResetsCount(t *testing.T) {
	// Use a very short window by creating a throttler and manually
	// waiting; we rely on the 1-second window being exceeded.
	// This test is intentionally skipped in short mode.
	if testing.Short() {
		t.Skip("skipping window expiry test in short mode")
	}

	th := throttle.New(throttle.WithRate(1))
	ctx := context.Background()

	_ = th.Allow(ctx, "key1")
	time.Sleep(1100 * time.Millisecond)

	if err := th.Allow(ctx, "key1"); err != nil {
		t.Fatalf("expected allow after window expiry, got: %v", err)
	}
}
