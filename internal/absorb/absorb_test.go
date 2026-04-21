package absorb_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/absorb"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupAbsorber(t *testing.T) (*absorb.Absorber, *provider.Registry) {
	t.Helper()
	reg := provider.NewRegistry()
	reg.Register("src", mock.New())
	reg.Register("dst", mock.New())
	ab, err := absorb.New(reg, "dst")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return ab, reg
}

func TestAbsorb_CopiesSecret(t *testing.T) {
	ab, reg := setupAbsorber(t)
	ctx := context.Background()
	src, _ := reg.Get("src")
	_ = src.Put(ctx, "api/key", "s3cr3t")

	if err := ab.Absorb(ctx, "src", "api/key"); err != nil {
		t.Fatalf("Absorb: %v", err)
	}
	dst, _ := reg.Get("dst")
	val, err := dst.Get(ctx, "api/key")
	if err != nil {
		t.Fatalf("Get from dst: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("want s3cr3t, got %q", val)
	}
}

func TestAbsorb_UnknownSourceProvider(t *testing.T) {
	ab, _ := setupAbsorber(t)
	err := ab.Absorb(context.Background(), "missing", "k")
	if err == nil {
		t.Fatal("expected error for unknown source provider")
	}
}

func TestAbsorb_MissingKey(t *testing.T) {
	ab, _ := setupAbsorber(t)
	err := ab.Absorb(context.Background(), "src", "no-such-key")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestAbsorbAll_RecordsResults(t *testing.T) {
	ab, reg := setupAbsorber(t)
	ctx := context.Background()
	src, _ := reg.Get("src")
	_ = src.Put(ctx, "x", "1")
	_ = src.Put(ctx, "y", "2")

	results := ab.AbsorbAll(ctx, "src")
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	if absorb.HasFailures(results) {
		t.Errorf("unexpected failures: %+v", results)
	}
}

func TestAbsorbAll_UnknownProvider_ReturnsError(t *testing.T) {
	ab, _ := setupAbsorber(t)
	results := ab.AbsorbAll(context.Background(), "ghost")
	if !absorb.HasFailures(results) {
		t.Fatal("expected failure for unknown provider")
	}
}

func TestNew_NilRegistry_ReturnsError(t *testing.T) {
	_, err := absorb.New(nil, "dst")
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}

func TestNew_EmptyDest_ReturnsError(t *testing.T) {
	_, err := absorb.New(provider.NewRegistry(), "")
	if err == nil {
		t.Fatal("expected error for empty dest")
	}
}
