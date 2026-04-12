package environ_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/environ"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupExporter(t *testing.T, opts ...environ.Option) (*environ.Exporter, *mock.Provider) {
	t.Helper()
	mp := mock.New()
	reg := provider.NewRegistry()
	reg.Register("test", mp)
	return environ.New(reg, opts...), mp
}

func TestRender_ReturnsKeyValueMap(t *testing.T) {
	exp, mp := setupExporter(t)
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "db_pass", "s3cr3t")
	_ = mp.PutSecret(ctx, "api_key", "abc123")

	m, err := exp.Render(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["db_pass"] != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %q", m["db_pass"])
	}
	if m["api_key"] != "abc123" {
		t.Errorf("expected abc123, got %q", m["api_key"])
	}
}

func TestRender_WithUpperCase(t *testing.T) {
	exp, mp := setupExporter(t, environ.WithUpperCase())
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "db_pass", "val")

	m, err := exp.Render(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m["DB_PASS"]; !ok {
		t.Errorf("expected upper-case key DB_PASS, got keys: %v", m)
	}
}

func TestRender_WithPrefix(t *testing.T) {
	exp, mp := setupExporter(t, environ.WithPrefix("APP_"))
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "token", "xyz")

	m, err := exp.Render(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["APP_token"] != "xyz" {
		t.Errorf("expected APP_token=xyz, got %v", m)
	}
}

func TestRender_WithQuoteValues(t *testing.T) {
	exp, mp := setupExporter(t, environ.WithQuoteValues())
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "pw", "hello")

	m, err := exp.Render(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["pw"] != `"hello"` {
		t.Errorf("expected quoted value, got %q", m["pw"])
	}
}

func TestRender_UnknownProvider(t *testing.T) {
	reg := provider.NewRegistry()
	exp := environ.New(reg)
	_, err := exp.Render(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestLines_ReturnsFlatSlice(t *testing.T) {
	exp, mp := setupExporter(t)
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "key1", "val1")

	lines, err := exp.Lines(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "key1=val1" {
		t.Errorf("unexpected line: %q", lines[0])
	}
}
