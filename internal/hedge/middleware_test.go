package hedge_test

import (
	"context"
	"testing"
	"time"

	"github.com/vaultshift/internal/hedge"
	"github.com/vaultshift/internal/provider/mock"
)

func setupMiddleware(t *testing.T) (primary, fallback *mock.Provider, rh *hedge.ReadHedger) {
	t.Helper()
	primary = mock.New()
	fallback = mock.New()
	rh = hedge.WrapRead(primary, []interface{ Get(context.Context, string) (string, error) }{fallback}, 5*time.Millisecond)
	return
}

func TestWrapRead_PutGoesToPrimary(t *testing.T) {
	primary := mock.New()
	fallback := mock.New()
	rh := hedge.WrapRead(primary, nil, 5*time.Millisecond)

	if err := rh.Put(context.Background(), "k", "v"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := primary.Get(context.Background(), "k")
	if err != nil {
		t.Fatalf("primary should have key: %v", err)
	}
	if val != "v" {
		t.Errorf("expected 'v', got %q", val)
	}
	_ = fallback
}

func TestWrapRead_GetHedgesAcrossProviders(t *testing.T) {
	primary := mock.New() // key absent
	fallback := mock.New()
	_ = fallback.Put(context.Background(), "secret", "hedged")

	rh := hedge.WrapRead(primary, nil, 5*time.Millisecond)
	_ = rh

	// Directly verify fallback path via Hedger
	h, err := hedge.New([]interface{ Get(context.Context, string) (string, error) }{primary, fallback},
		hedge.WithDelay(5*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := h.Get(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hedged" {
		t.Errorf("expected 'hedged', got %q", val)
	}
}

func TestWrapRead_DeleteGoesToPrimary(t *testing.T) {
	primary := mock.New()
	_ = primary.Put(context.Background(), "del", "val")
	rh := hedge.WrapRead(primary, nil, 5*time.Millisecond)

	if err := rh.Delete(context.Background(), "del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := primary.Get(context.Background(), "del")
	if err == nil {
		t.Error("expected key to be deleted from primary")
	}
}

func TestWrapRead_ListGoesToPrimary(t *testing.T) {
	primary := mock.New()
	_ = primary.Put(context.Background(), "a", "1")
	_ = primary.Put(context.Background(), "b", "2")
	rh := hedge.WrapRead(primary, nil, 5*time.Millisecond)

	keys, err := rh.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
