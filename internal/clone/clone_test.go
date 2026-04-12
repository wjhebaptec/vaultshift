package clone_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/clone"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupCloner(t *testing.T) (*clone.Cloner, *mock.Provider, *mock.Provider) {
	t.Helper()
	src := mock.New()
	dst := mock.New()
	reg := provider.NewRegistry()
	reg.Register("src", src)
	reg.Register("dst", dst)
	return clone.New(reg), src, dst
}

func TestClone_CopiesSecret(t *testing.T) {
	cloner, src, dst := setupCloner(t)
	ctx := context.Background()

	_ = src.PutSecret(ctx, "db/password", "s3cr3t")

	res := cloner.Clone(ctx, "src", "dst", "db/password", clone.Options{})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	got, err := dst.GetSecret(ctx, "db/password")
	if err != nil {
		t.Fatalf("get from dst: %v", err)
	}
	if got != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", got)
	}
}

func TestClone_WithKeyTransform(t *testing.T) {
	cloner, src, dst := setupCloner(t)
	ctx := context.Background()

	_ = src.PutSecret(ctx, "foo", "bar")

	opts := clone.Options{KeyTransform: func(k string) string { return "prod/" + k }}
	res := cloner.Clone(ctx, "src", "dst", "foo", opts)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.DestKey != "prod/foo" {
		t.Errorf("expected dest key %q, got %q", "prod/foo", res.DestKey)
	}

	got, err := dst.GetSecret(ctx, "prod/foo")
	if err != nil || got != "bar" {
		t.Errorf("expected %q at prod/foo, got %q (err %v)", "bar", got, err)
	}
}

func TestClone_DryRun_DoesNotWrite(t *testing.T) {
	cloner, src, dst := setupCloner(t)
	ctx := context.Background()

	_ = src.PutSecret(ctx, "key", "value")

	res := cloner.Clone(ctx, "src", "dst", "key", clone.Options{DryRun: true})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Skipped {
		t.Error("expected result to be skipped in dry-run mode")
	}

	_, err := dst.GetSecret(ctx, "key")
	if err == nil {
		t.Error("expected key to be absent in dst after dry-run")
	}
}

func TestClone_UnknownSourceProvider(t *testing.T) {
	cloner, _, _ := setupCloner(t)
	res := cloner.Clone(context.Background(), "missing", "dst", "k", clone.Options{})
	if res.Err == nil {
		t.Error("expected error for unknown source provider")
	}
}

func TestCloneAll_CopiesAllKeys(t *testing.T) {
	cloner, src, dst := setupCloner(t)
	ctx := context.Background()

	_ = src.PutSecret(ctx, "a", "1")
	_ = src.PutSecret(ctx, "b", "2")

	results := cloner.CloneAll(ctx, "src", "dst", clone.Options{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for key %q: %v", r.SourceKey, r.Err)
		}
	}

	keys, _ := dst.ListSecrets(ctx)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys in dst, got %d", len(keys))
	}
}
