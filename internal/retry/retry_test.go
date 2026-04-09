package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultshift/internal/retry"
)

var errTemp = errors.New("temporary error")

func TestDo_SucceedsFirstAttempt(t *testing.T) {
	r := retry.New(retry.DefaultConfig())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	cfg := retry.Config{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, Multiplier: 2.0}
	r := retry.New(cfg)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errTemp
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	cfg := retry.Config{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, Multiplier: 1.5}
	r := retry.New(cfg)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return errTemp
	})
	if !errors.Is(err, retry.ErrMaxAttemptsReached) {
		t.Fatalf("expected ErrMaxAttemptsReached, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	cfg := retry.Config{MaxAttempts: 5, InitialDelay: 50 * time.Millisecond, MaxDelay: time.Second, Multiplier: 2.0}
	r := retry.New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := r.Do(ctx, func() error {
		calls++
		cancel()
		return errTemp
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestNew_InvalidMaxAttempts_DefaultsToOne(t *testing.T) {
	r := retry.New(retry.Config{MaxAttempts: 0, InitialDelay: time.Millisecond, Multiplier: 2.0})
	calls := 0
	r.Do(context.Background(), func() error { //nolint:errcheck
		calls++
		return errTemp
	})
	if calls != 1 {
		t.Fatalf("expected exactly 1 call for MaxAttempts=0, got %d", calls)
	}
}
