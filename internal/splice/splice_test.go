package splice_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/splice"
)

func setupSplicer(t *testing.T, rewrite func(string, string) string) (*splice.Splicer, *mock.Provider, *mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	src := mock.New()
	dst := mock.New()
	reg.Register("src", src)
	reg.Register("dst", dst)
	s, err := splice.New(reg, "dst", rewrite)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, src, dst
}

func TestSplice_CopiesSecret(t *testing.T) {
	s, src, dst := setupSplicer(t, nil)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "db/pass", "s3cr3t")
	if err := s.Splice(ctx, "src", "db/pass"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := dst.GetSecret(ctx, "db/pass")
	if err != nil || val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %q (%v)", val, err)
	}
}

func TestSplice_WithRewrite(t *testing.T) {
	rewrite := func(src, key string) string { return strings.ToUpper(src) + "/" + key }
	s, src, dst := setupSplicer(t, rewrite)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "token", "abc123")
	if err := s.Splice(ctx, "src", "token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := dst.GetSecret(ctx, "SRC/token")
	if err != nil || val != "abc123" {
		t.Fatalf("expected abc123 at SRC/token, got %q (%v)", val, err)
	}
}

func TestSplice_UnknownSourceProvider(t *testing.T) {
	s, _, _ := setupSplicer(t, nil)
	err := s.Splice(context.Background(), "missing", "key")
	if err == nil {
		t.Fatal("expected error for unknown source provider")
	}
}

func TestSpliceAll_RecordsFailures(t *testing.T) {
	s, src, _ := setupSplicer(t, nil)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "a", "1")
	s.SpliceAll(ctx, "src", []string{"a", "missing-key"})
	results := s.Results()
	if len(results) < 1 {
		t.Fatal("expected at least one result")
	}
}

func TestHasFailures_False(t *testing.T) {
	s, src, _ := setupSplicer(t, nil)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "x", "y")
	s.SpliceAll(ctx, "src", []string{"x"})
	if s.HasFailures() {
		t.Fatal("expected no failures")
	}
}

func TestNew_NilRegistry_ReturnsError(t *testing.T) {
	_, err := splice.New(nil, "dst", nil)
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}

func TestNew_EmptyDest_ReturnsError(t *testing.T) {
	_, err := splice.New(provider.NewRegistry(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty dest provider")
	}
}
