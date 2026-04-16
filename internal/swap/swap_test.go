package swap_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/swap"
)

func setupSwapper(t *testing.T, bidirectional bool) (*swap.Swapper, *provider.Registry) {
	t.Helper()
	reg := provider.NewRegistry()
	reg.Register("alpha", mock.New())
	reg.Register("beta", mock.New())
	s, err := swap.New(reg, bidirectional)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, reg
}

func TestSwap_MovesValueToDest(t *testing.T) {
	s, reg := setupSwapper(t, false)
	ctx := context.Background()
	alpha, _ := reg.Get("alpha")
	_ = alpha.Put(ctx, "db/pass", "secret123")

	res := s.Swap(ctx, "alpha", "beta", "db/pass")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Swapped {
		t.Fatal("expected Swapped=true")
	}
	beta, _ := reg.Get("beta")
	got, err := beta.Get(ctx, "db/pass")
	if err != nil || got != "secret123" {
		t.Fatalf("expected secret123, got %q (%v)", got, err)
	}
}

func TestSwap_Bidirectional_ExchangesValues(t *testing.T) {
	s, reg := setupSwapper(t, true)
	ctx := context.Background()
	alpha, _ := reg.Get("alpha")
	beta, _ := reg.Get("beta")
	_ = alpha.Put(ctx, "key", "from-alpha")
	_ = beta.Put(ctx, "key", "from-beta")

	res := s.Swap(ctx, "alpha", "beta", "key")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	gotBeta, _ := beta.Get(ctx, "key")
	gotAlpha, _ := alpha.Get(ctx, "key")
	if gotBeta != "from-alpha" {
		t.Errorf("beta: want from-alpha, got %q", gotBeta)
	}
	if gotAlpha != "from-beta" {
		t.Errorf("alpha: want from-beta, got %q", gotAlpha)
	}
}

func TestSwap_UnknownSourceProvider(t *testing.T) {
	s, _ := setupSwapper(t, false)
	res := s.Swap(context.Background(), "missing", "beta", "k")
	if res.Err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestSwap_UnknownDestProvider(t *testing.T) {
	s, _ := setupSwapper(t, false)
	res := s.Swap(context.Background(), "alpha", "missing", "k")
	if res.Err == nil {
		t.Fatal("expected error for unknown dest")
	}
}

func TestSwapAll_RecordsFailures(t *testing.T) {
	s, reg := setupSwapper(t, false)
	ctx := context.Background()
	alpha, _ := reg.Get("alpha")
	_ = alpha.Put(ctx, "a", "val-a")

	results := s.SwapAll(ctx, "alpha", "beta", []string{"a", "nonexistent"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("first swap should succeed: %v", results[0].Err)
	}
}

func TestHasFailures_True(t *testing.T) {
	s, _ := setupSwapper(t, false)
	results := s.SwapAll(context.Background(), "alpha", "beta", []string{"missing"})
	if !swap.HasFailures(results) {
		t.Error("expected HasFailures=true")
	}
}

func TestNew_NilRegistry_ReturnsError(t *testing.T) {
	_, err := swap.New(nil, false)
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}
