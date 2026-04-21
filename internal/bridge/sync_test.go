package bridge_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/bridge"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func TestSync_BidirectionalCopies(t *testing.T) {
	reg := provider.NewRegistry()
	a := mock.New()
	b := mock.New()
	reg.Register("a", a)
	reg.Register("b", b)

	ctx := context.Background()
	_ = a.PutSecret(ctx, "from-a", "val-a")
	_ = b.PutSecret(ctx, "from-b", "val-b")

	res, err := bridge.Sync(ctx, reg, "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.HasFailures() {
		t.Fatal("expected no failures")
	}

	v, err := b.GetSecret(ctx, "from-a")
	if err != nil || v != "val-a" {
		t.Errorf("expected val-a in b, got %q err=%v", v, err)
	}
	v, err = a.GetSecret(ctx, "from-b")
	if err != nil || v != "val-b" {
		t.Errorf("expected val-b in a, got %q err=%v", v, err)
	}
}

func TestSync_UnknownProviderA_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register("b", mock.New())
	_, err := bridge.Sync(context.Background(), reg, "missing", "b")
	if err == nil {
		t.Fatal("expected error for unknown provider A")
	}
}

func TestSync_UnknownProviderB_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register("a", mock.New())
	_, err := bridge.Sync(context.Background(), reg, "a", "missing")
	if err == nil {
		t.Fatal("expected error for unknown provider B")
	}
}

func TestSync_EmptyProviders_NoResults(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register("a", mock.New())
	reg.Register("b", mock.New())

	res, err := bridge.Sync(context.Background(), reg, "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.AtoB) != 0 || len(res.BtoA) != 0 {
		t.Errorf("expected empty results, got AtoB=%d BtoA=%d", len(res.AtoB), len(res.BtoA))
	}
}

func TestSyncResult_HasFailures_False(t *testing.T) {
	res := bridge.SyncResult{
		AtoB: []bridge.Result{{Key: "k"}},
		BtoA: []bridge.Result{{Key: "j"}},
	}
	if res.HasFailures() {
		t.Error("expected no failures")
	}
}
