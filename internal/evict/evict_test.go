package evict_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/evict"
	"github.com/vaultshift/internal/provider/mock"
)

func setup(t *testing.T, cap int) (*evict.Cache, *mock.Provider) {
	t.Helper()
	m := mock.New()
	c, err := evict.New(m, cap)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, m
}

func TestNew_InvalidCapacity_ReturnsError(t *testing.T) {
	_, err := evict.New(mock.New(), 0)
	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestNew_NilProvider_ReturnsError(t *testing.T) {
	_, err := evict.New(nil, 5)
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestPutAndGet_CachesValue(t *testing.T) {
	c, _ := setup(t, 3)
	ctx := context.Background()

	if err := c.Put(ctx, "k1", "v1"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	val, err := c.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "v1" {
		t.Errorf("got %q, want %q", val, "v1")
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d, want 1", c.Len())
	}
}

func TestGet_FallsBackToBackend(t *testing.T) {
	c, m := setup(t, 3)
	ctx := context.Background()

	_ = m.Put(ctx, "remote", "secret")
	val, err := c.Get(ctx, "remote")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "secret" {
		t.Errorf("got %q, want %q", val, "secret")
	}
}

func TestEvict_LRUEntryRemoved(t *testing.T) {
	c, _ := setup(t, 2)
	ctx := context.Background()

	_ = c.Put(ctx, "a", "1")
	_ = c.Put(ctx, "b", "2")
	// Access "a" so "b" becomes LRU
	_, _ = c.Get(ctx, "a")
	// Insert "c" — "b" should be evicted
	_ = c.Put(ctx, "c", "3")

	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
}

func TestDelete_RemovesFromCacheAndBackend(t *testing.T) {
	c, m := setup(t, 3)
	ctx := context.Background()

	_ = c.Put(ctx, "x", "y")
	if err := c.Delete(ctx, "x"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}
	if _, err := m.Get(ctx, "x"); err == nil {
		t.Error("expected backend to have no entry for x")
	}
}

func TestList_DelegatesToBackend(t *testing.T) {
	c, m := setup(t, 3)
	ctx := context.Background()

	_ = m.Put(ctx, "p", "1")
	_ = m.Put(ctx, "q", "2")
	keys, err := c.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("len(keys) = %d, want 2", len(keys))
	}
}
