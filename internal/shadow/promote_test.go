package shadow_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/shadow"
)

func TestPromote_CopiesAllShadowKeysToPrimary(t *testing.T) {
	prim := mock.New()
	shad := mock.New()
	ctx := context.Background()

	_ = shad.Put(ctx, "a", "1")
	_ = shad.Put(ctx, "b", "2")

	s, _ := shadow.New(prim, shad, shadow.ModeCompare)
	results, err := s.Promote(ctx)
	if err != nil {
		t.Fatalf("promote failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("key %q: unexpected error: %v", r.Key, r.Err)
		}
	}

	v, _ := prim.Get(ctx, "a")
	if v != "1" {
		t.Errorf("primary[a]: got %q, want %q", v, "1")
	}
	v, _ = prim.Get(ctx, "b")
	if v != "2" {
		t.Errorf("primary[b]: got %q, want %q", v, "2")
	}
}

func TestPromote_EmptyShadow_ReturnsNoResults(t *testing.T) {
	prim := mock.New()
	shad := mock.New()
	s, _ := shadow.New(prim, shad, shadow.ModeWriteOnly)

	results, err := s.Promote(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
