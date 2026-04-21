package bounce_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/bounce"
	"github.com/vaultshift/internal/provider/mock"
)

func setup(t *testing.T, batchSize int) (*bounce.Buffer, *mock.Provider) {
	t.Helper()
	mp := mock.New()
	b, err := bounce.New(mp, batchSize)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return b, mp
}

func TestNew_NilProvider_ReturnsError(t *testing.T) {
	_, err := bounce.New(nil, 5)
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestNew_InvalidBatchSize_ReturnsError(t *testing.T) {
	mp := mock.New()
	_, err := bounce.New(mp, 0)
	if err == nil {
		t.Fatal("expected error for batchSize=0")
	}
}

func TestPut_BuffersWithoutFlush(t *testing.T) {
	b, mp := setup(t, 5)
	ctx := context.Background()

	_ = b.Put(ctx, "k1", "v1")
	_ = b.Put(ctx, "k2", "v2")

	if b.Len() != 2 {
		t.Fatalf("expected 2 buffered entries, got %d", b.Len())
	}
	keys, _ := mp.List(ctx)
	if len(keys) != 0 {
		t.Fatalf("expected provider to be empty before flush, got %d keys", len(keys))
	}
}

func TestPut_AutoFlushesAtBatchSize(t *testing.T) {
	b, mp := setup(t, 3)
	ctx := context.Background()

	_ = b.Put(ctx, "a", "1")
	_ = b.Put(ctx, "b", "2")
	if err := b.Put(ctx, "c", "3"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if b.Len() != 0 {
		t.Fatalf("expected buffer to be empty after auto-flush, got %d", b.Len())
	}
	keys, _ := mp.List(ctx)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys in provider, got %d", len(keys))
	}
}

func TestFlush_WritesAllEntries(t *testing.T) {
	b, mp := setup(t, 10)
	ctx := context.Background()

	_ = b.Put(ctx, "x", "hello")
	_ = b.Put(ctx, "y", "world")

	results := b.Flush(ctx)
	for _, r := range results {
		if r.Error != nil {
			t.Fatalf("flush error for key %q: %v", r.Key, r.Error)
		}
	}

	v, err := mp.Get(ctx, "x")
	if err != nil || v != "hello" {
		t.Fatalf("expected x=hello, got %q %v", v, err)
	}
}

func TestFlush_ClearsBuffer(t *testing.T) {
	b, _ := setup(t, 10)
	ctx := context.Background()

	_ = b.Put(ctx, "z", "val")
	b.Flush(ctx)

	if b.Len() != 0 {
		t.Fatalf("expected empty buffer after flush, got %d", b.Len())
	}
}

func TestFlush_ReturnsErrorResults(t *testing.T) {
	mp := mock.New()
	mp.Close()
	b, _ := bounce.New(mp, 10)
	ctx := context.Background()

	// Bypass auto-flush by manipulating via Flush directly after manual inject.
	// We call Put on the closed provider to trigger an error path.
	err := b.Put(ctx, "bad", "val")
	// Put itself won't flush (batchSz=10, only 1 entry), so manually flush.
	results := b.Flush(ctx)
	_ = err

	for _, r := range results {
		if r.Error == nil {
			t.Fatalf("expected error for key %q on closed provider", r.Key)
		}
		if !errors.Is(r.Error, mock.ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", r.Error)
		}
	}
}
