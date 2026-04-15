package evict_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/evict"
	"github.com/vaultshift/internal/provider/mock"
)

func TestWrap_ReturnsProvider(t *testing.T) {
	m := mock.New()
	p, err := evict.Wrap(m, 4)
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestWrap_InvalidCapacity_ReturnsError(t *testing.T) {
	_, err := evict.Wrap(mock.New(), -1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrap_PutAndGet_RoundTrip(t *testing.T) {
	ctx := context.Background()
	p, _ := evict.Wrap(mock.New(), 5)

	if err := p.Put(ctx, "key", "val"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := p.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "val" {
		t.Errorf("got %q, want %q", got, "val")
	}
}

func TestWrap_Delete_RemovesKey(t *testing.T) {
	ctx := context.Background()
	p, _ := evict.Wrap(mock.New(), 5)

	_ = p.Put(ctx, "del", "gone")
	if err := p.Delete(ctx, "del"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := p.Get(ctx, "del"); err == nil {
		t.Error("expected error after delete")
	}
}

func TestWrap_List_ReturnsKeys(t *testing.T) {
	ctx := context.Background()
	p, _ := evict.Wrap(mock.New(), 5)

	_ = p.Put(ctx, "a", "1")
	_ = p.Put(ctx, "b", "2")
	keys, err := p.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("len = %d, want 2", len(keys))
	}
}
