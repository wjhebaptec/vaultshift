package propagate_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/propagate"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupPropagator(t *testing.T) (*propagate.Propagator, *mock.Provider, *mock.Provider) {
	t.Helper()
	src := mock.New("src")
	dst := mock.New("dst")
	reg := provider.NewRegistry()
	reg.Register(src)
	reg.Register(dst)
	return propagate.New(reg), src, dst
}

func TestPropagate_CopiesSecret(t *testing.T) {
	p, src, dst := setupPropagator(t)
	ctx := context.Background()

	_ = src.PutSecret(ctx, "db/password", "s3cr3t")
	p.AddRule(propagate.Rule{
		SourceProvider: "src",
		SourceKey:      "db/password",
		DestProvider:   "dst",
		DestKey:        "db/password",
	})

	if err := p.Propagate(ctx); err != nil {
		t.Fatalf("Propagate: %v", err)
	}
	got, err := dst.GetSecret(ctx, "db/password")
	if err != nil || got != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %q (%v)", got, err)
	}
}

func TestPropagate_RenamesKey(t *testing.T) {
	p, src, dst := setupPropagator(t)
	ctx := context.Background()

	_ = src.PutSecret(ctx, "old/key", "value123")
	p.AddRule(propagate.Rule{
		SourceProvider: "src", SourceKey: "old/key",
		DestProvider: "dst", DestKey: "new/key",
	})

	if err := p.Propagate(ctx); err != nil {
		t.Fatalf("Propagate: %v", err)
	}
	got, _ := dst.GetSecret(ctx, "new/key")
	if got != "value123" {
		t.Fatalf("expected value123, got %q", got)
	}
}

func TestPropagate_UnknownSourceProvider(t *testing.T) {
	p, _, _ := setupPropagator(t)
	p.AddRule(propagate.Rule{
		SourceProvider: "ghost", SourceKey: "k",
		DestProvider: "dst", DestKey: "k",
	})
	if err := p.Propagate(context.Background()); err == nil {
		t.Fatal("expected error for unknown source provider")
	}
}

func TestPropagate_UnknownDestProvider(t *testing.T) {
	p, src, _ := setupPropagator(t)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "k", "v")
	p.AddRule(propagate.Rule{
		SourceProvider: "src", SourceKey: "k",
		DestProvider: "ghost", DestKey: "k",
	})
	if err := p.Propagate(ctx); err == nil {
		t.Fatal("expected error for unknown dest provider")
	}
}

func TestPropagateAll_CollectsErrors(t *testing.T) {
	p, src, _ := setupPropagator(t)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "k", "v")

	p.AddRule(propagate.Rule{SourceProvider: "src", SourceKey: "k", DestProvider: "dst", DestKey: "k"})
	p.AddRule(propagate.Rule{SourceProvider: "ghost", SourceKey: "x", DestProvider: "dst", DestKey: "x"})

	if err := p.PropagateAll(ctx); err == nil {
		t.Fatal("expected combined error")
	}
}
