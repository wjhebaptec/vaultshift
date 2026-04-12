package relay_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/relay"
)

func setupFilteredRelay(t *testing.T) (*relay.Relay, *mock.Provider, *mock.Provider) {
	t.Helper()
	src := mock.New()
	dst := mock.New()
	reg := provider.NewRegistry()
	reg.Register("src", src)
	reg.Register("dst", dst)
	return relay.New(reg), src, dst
}

func TestWithSourcePrefix_Matches(t *testing.T) {
	rl, src, dst := setupFilteredRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "prod/key", "v1")
	_ = src.Put(ctx, "dev/key", "v2")
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "prod/key", DestProvider: "dst"})
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "dev/key", DestProvider: "dst"})
	fr := relay.NewFiltered(rl, relay.WithSourcePrefix("prod/"))
	results := fr.Forward(ctx)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Rule.SourceKey != "prod/key" {
		t.Fatalf("unexpected key %q", results[0].Rule.SourceKey)
	}
}

func TestWithDestProvider_Filters(t *testing.T) {
	rl, src, _ := setupFilteredRelay(t)
	dst2 := mock.New()
	ctx := context.Background()
	_ = src.Put(ctx, "k1", "a")
	_ = src.Put(ctx, "k2", "b")
	rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "k1", DestProvider: "dst"})
	rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "k2", DestProvider: "dst2"})
	_ = dst2 // registered separately if needed
	fr := relay.NewFiltered(rl, relay.WithDestProvider("dst"))
	results := fr.Forward(ctx)
	if len(results) != 1 || results[0].Rule.DestProvider != "dst" {
		t.Fatalf("expected only dst rule, got %+v", results)
	}
}

func TestChainFilters_BothMustPass(t *testing.T) {
	rl, src, _ := setupFilteredRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "prod/api", "val")
	_ = src.Put(ctx, "dev/api", "val2")
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "prod/api", DestProvider: "dst"})
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "dev/api", DestProvider: "dst"})
	chained := relay.ChainFilters(
		relay.WithSourcePrefix("prod/"),
		relay.WithDestProvider("dst"),
	)
	fr := relay.NewFiltered(rl, chained)
	results := fr.Forward(ctx)
	if len(results) != 1 || results[0].Rule.SourceKey != "prod/api" {
		t.Fatalf("expected only prod/api, got %+v", results)
	}
}

func TestChainFilters_NoFilters_AllPass(t *testing.T) {
	rl, src, _ := setupFilteredRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "a", "1")
	_ = src.Put(ctx, "b", "2")
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "a", DestProvider: "dst"})
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "b", DestProvider: "dst"})
	fr := relay.NewFiltered(rl, relay.ChainFilters())
	results := fr.Forward(ctx)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
