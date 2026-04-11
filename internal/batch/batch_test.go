package batch_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/vaultshift/internal/batch"
)

func makeItems(keys ...string) []batch.Item {
	items := make([]batch.Item, len(keys))
	for i, k := range keys {
		items[i] = batch.Item{Key: k, Value: "v" + k}
	}
	return items
}

func TestRun_AllSucceed(t *testing.T) {
	p := batch.New(batch.WithSize(2), batch.WithWorkers(2))
	items := makeItems("a", "b", "c", "d")
	results := p.Run(context.Background(), items, func(_ context.Context, item batch.Item) error {
		return nil
	})
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for key %q: %v", r.Key, r.Err)
		}
	}
}

func TestRun_PartialFailure(t *testing.T) {
	p := batch.New(batch.WithSize(3))
	items := makeItems("x", "y", "z")
	sentinel := errors.New("fail")
	results := p.Run(context.Background(), items, func(_ context.Context, item batch.Item) error {
		if item.Key == "y" {
			return sentinel
		}
		return nil
	})
	var failed int
	for _, r := range results {
		if r.Err != nil {
			failed++
			if !errors.Is(r.Err, sentinel) {
				t.Errorf("unexpected error: %v", r.Err)
			}
		}
	}
	if failed != 1 {
		t.Errorf("expected 1 failure, got %d", failed)
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	p := batch.New(batch.WithSize(5))
	items := makeItems("a", "b", "c")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := p.Run(ctx, items, func(_ context.Context, _ batch.Item) error {
		return nil
	})
	for _, r := range results {
		if r.Err == nil {
			t.Errorf("expected cancellation error for key %q", r.Key)
		}
	}
}

func TestRun_WorkerConcurrency(t *testing.T) {
	p := batch.New(batch.WithSize(10), batch.WithWorkers(4))
	items := makeItems("1", "2", "3", "4", "5", "6", "7", "8")
	var count int64
	p.Run(context.Background(), items, func(_ context.Context, _ batch.Item) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	if int(atomic.LoadInt64(&count)) != len(items) {
		t.Errorf("expected %d invocations, got %d", len(items), count)
	}
}

func TestRun_EmptyItems(t *testing.T) {
	p := batch.New()
	results := p.Run(context.Background(), nil, func(_ context.Context, _ batch.Item) error {
		t.Fatal("fn should not be called for empty input")
		return nil
	})
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
