package bridge_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/bridge"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupBridge(t *testing.T) (*provider.Registry, *mock.Provider, *mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	src := mock.New()
	dst := mock.New()
	reg.Register("src", src)
	reg.Register("dst", dst)
	return reg, src, dst
}

func TestNew_NilRegistry_ReturnsError(t *testing.T) {
	_, err := bridge.New(nil, "src", "dst")
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}

func TestNew_EmptySource_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	_, err := bridge.New(reg, "", "dst")
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestNew_EmptyDest_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	_, err := bridge.New(reg, "src", "")
	if err == nil {
		t.Fatal("expected error for empty destination")
	}
}

func TestForward_CopiesAllKeys(t *testing.T) {
	reg, src, dst := setupBridge(t)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "alpha", "1")
	_ = src.PutSecret(ctx, "beta", "2")

	b, _ := bridge.New(reg, "src", "dst")
	results, err := b.Forward(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if bridge.HasFailures(results) {
		t.Fatal("expected no failures")
	}
	v, _ := dst.GetSecret(ctx, "alpha")
	if v != "1" {
		t.Errorf("expected '1', got %q", v)
	}
}

func TestForward_WithTransform_RenamesKey(t *testing.T) {
	reg, src, dst := setupBridge(t)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "key", "val")

	b, _ := bridge.New(reg, "src", "dst")
	b.WithTransform(func(k string) string { return "pfx_" + k })
	_, err := b.Forward(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, err := dst.GetSecret(ctx, "pfx_key")
	if err != nil || v != "val" {
		t.Errorf("expected 'val' at transformed key, got %q err=%v", v, err)
	}
}

func TestForward_WithFilter_SkipsNonMatching(t *testing.T) {
	reg, src, dst := setupBridge(t)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "keep", "yes")
	_ = src.PutSecret(ctx, "drop", "no")

	b, _ := bridge.New(reg, "src", "dst")
	b.WithFilter(func(k string) bool { return k == "keep" })
	results, _ := b.Forward(ctx)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	_, err := dst.GetSecret(ctx, "drop")
	if err == nil {
		t.Error("expected 'drop' to not be forwarded")
	}
}

func TestForward_UnknownSourceProvider_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	b, _ := bridge.New(reg, "missing", "dst")
	_, err := b.Forward(context.Background())
	if err == nil {
		t.Fatal("expected error for unknown source provider")
	}
}

func TestForward_DestWriteError_RecordedInResults(t *testing.T) {
	reg, src, _ := setupBridge(t)
	ctx := context.Background()
	_ = src.PutSecret(ctx, "k", "v")

	fail := &failProvider{err: errors.New("write error")}
	reg.Register("fail", fail)

	b, _ := bridge.New(reg, "src", "fail")
	results, err := b.Forward(ctx)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if !bridge.HasFailures(results) {
		t.Fatal("expected failure in results")
	}
}

type failProvider struct{ err error }

func (f *failProvider) GetSecret(_ context.Context, _ string) (string, error) {
	return "", f.err
}
func (f *failProvider) PutSecret(_ context.Context, _, _ string) error { return f.err }
func (f *failProvider) DeleteSecret(_ context.Context, _ string) error { return f.err }
func (f *failProvider) ListSecrets(_ context.Context) ([]string, error) {
	return nil, f.err
}
func (f *failProvider) Close() error { return nil }
