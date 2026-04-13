package stagger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultshift/internal/stagger"
)

func TestAdd_ValidTask(t *testing.T) {
	r := stagger.New()
	if err := r.Add("task1", func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdd_EmptyName_ReturnsError(t *testing.T) {
	r := stagger.New()
	if err := r.Add("", func(ctx context.Context) error { return nil }); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestAdd_NilFn_ReturnsError(t *testing.T) {
	r := stagger.New()
	if err := r.Add("task1", nil); err == nil {
		t.Fatal("expected error for nil fn")
	}
}

func TestRun_AllSucceed(t *testing.T) {
	r := stagger.New(stagger.WithDelay(0))
	called := make([]string, 0)
	for _, name := range []string{"a", "b", "c"} {
		n := name
		_ = r.Add(n, func(ctx context.Context) error {
			called = append(called, n)
			return nil
		})
	}
	results := r.Run(context.Background())
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, res := range results {
		if res.Err != nil {
			t.Errorf("unexpected error for %s: %v", res.Name, res.Err)
		}
	}
	if stagger.HasFailures(results) {
		t.Error("HasFailures should be false")
	}
}

func TestRun_PartialFailure(t *testing.T) {
	r := stagger.New(stagger.WithDelay(0))
	errBoom := errors.New("boom")
	_ = r.Add("ok", func(ctx context.Context) error { return nil })
	_ = r.Add("fail", func(ctx context.Context) error { return errBoom })
	results := r.Run(context.Background())
	if !stagger.HasFailures(results) {
		t.Error("HasFailures should be true")
	}
	if results[1].Err != errBoom {
		t.Errorf("expected errBoom, got %v", results[1].Err)
	}
}

func TestRun_ContextCancelled_StopsEarly(t *testing.T) {
	r := stagger.New(stagger.WithDelay(10 * time.Millisecond))
	_ = r.Add("first", func(ctx context.Context) error { return nil })
	_ = r.Add("second", func(ctx context.Context) error { return nil })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := r.Run(ctx)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[1].Err == nil {
		t.Error("expected context error on second task")
	}
}

func TestRun_ElapsedIsRecorded(t *testing.T) {
	r := stagger.New(stagger.WithDelay(0))
	_ = r.Add("slow", func(ctx context.Context) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})
	results := r.Run(context.Background())
	if results[0].Elapsed < 5*time.Millisecond {
		t.Errorf("expected elapsed >= 5ms, got %v", results[0].Elapsed)
	}
}

func TestHasFailures_False_WhenEmpty(t *testing.T) {
	if stagger.HasFailures(nil) {
		t.Error("HasFailures should be false for nil slice")
	}
}
