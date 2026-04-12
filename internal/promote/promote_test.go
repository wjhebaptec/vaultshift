package promote_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/promote"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupPromoter(t *testing.T, opts promote.Options) (*promote.Promoter, *mock.Provider, *mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	src := mock.New()
	dst := mock.New()
	reg.Register("staging", src)
	reg.Register("prod", dst)
	return promote.New(reg, opts), src, dst
}

func TestPromote_CopiesSecret(t *testing.T) {
	p, src, dst := setupPromoter(t, promote.Options{})
	ctx := context.Background()
	_ = src.Put(ctx, "DB_PASS", "s3cr3t")

	r := p.Promote(ctx, "staging", "prod", "DB_PASS")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.DestKey != "DB_PASS" {
		t.Errorf("dest key = %q, want %q", r.DestKey, "DB_PASS")
	}
	val, _ := dst.Get(ctx, "DB_PASS")
	if val != "s3cr3t" {
		t.Errorf("value = %q, want %q", val, "s3cr3t")
	}
}

func TestPromote_WithKeyTransform(t *testing.T) {
	opts := promote.Options{
		KeyTransform: func(k string) string { return "prod/" + k },
	}
	p, src, dst := setupPromoter(t, opts)
	ctx := context.Background()
	_ = src.Put(ctx, "API_KEY", "abc123")

	r := p.Promote(ctx, "staging", "prod", "API_KEY")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.DestKey != "prod/API_KEY" {
		t.Errorf("dest key = %q, want prod/API_KEY", r.DestKey)
	}
	val, _ := dst.Get(ctx, "prod/API_KEY")
	if val != "abc123" {
		t.Errorf("value = %q, want abc123", val)
	}
}

func TestPromote_DryRun_DoesNotWrite(t *testing.T) {
	p, src, dst := setupPromoter(t, promote.Options{DryRun: true})
	ctx := context.Background()
	_ = src.Put(ctx, "TOKEN", "xyz")

	r := p.Promote(ctx, "staging", "prod", "TOKEN")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if !r.DryRun {
		t.Error("expected DryRun=true in result")
	}
	keys, _ := dst.List(ctx)
	if len(keys) != 0 {
		t.Errorf("expected no keys written in dry-run, got %v", keys)
	}
}

func TestPromote_UnknownSourceProvider(t *testing.T) {
	p, _, _ := setupPromoter(t, promote.Options{})
	r := p.Promote(context.Background(), "ghost", "prod", "K")
	if r.Err == nil {
		t.Fatal("expected error for unknown source provider")
	}
}

func TestPromote_UnknownDestProvider(t *testing.T) {
	p, src, _ := setupPromoter(t, promote.Options{})
	ctx := context.Background()
	_ = src.Put(ctx, "K", "v")
	r := p.Promote(ctx, "staging", "nowhere", "K")
	if r.Err == nil {
		t.Fatal("expected error for unknown destination provider")
	}
}

func TestPromoteAll_CopiesAllKeys(t *testing.T) {
	p, src, dst := setupPromoter(t, promote.Options{})
	ctx := context.Background()
	_ = src.Put(ctx, "A", "1")
	_ = src.Put(ctx, "B", "2")

	results := p.PromoteAll(ctx, "staging", "prod")
	if promote.HasFailures(results) {
		t.Fatalf("unexpected failures: %+v", results)
	}
	keys, _ := dst.List(ctx)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys in dest, got %d", len(keys))
	}
}
