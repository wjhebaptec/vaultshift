package environ_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/vaultshift/internal/environ"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupWriter(t *testing.T, buf *bytes.Buffer, opts ...environ.Option) (*environ.Writer, *mock.Provider) {
	t.Helper()
	mp := mock.New()
	reg := provider.NewRegistry()
	reg.Register("test", mp)
	exp := environ.New(reg, opts...)
	return environ.NewWriter(exp, buf), mp
}

func TestWrite_OutputsSortedLines(t *testing.T) {
	var buf bytes.Buffer
	w, mp := setupWriter(t, &buf)
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "z_key", "last")
	_ = mp.PutSecret(ctx, "a_key", "first")

	if err := w.Write(ctx, "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "a_key=first" {
		t.Errorf("expected a_key=first first, got %q", lines[0])
	}
	if lines[1] != "z_key=last" {
		t.Errorf("expected z_key=last second, got %q", lines[1])
	}
}

func TestWrite_WithUpperCaseAndPrefix(t *testing.T) {
	var buf bytes.Buffer
	w, mp := setupWriter(t, &buf, environ.WithUpperCase(), environ.WithPrefix("MY_"))
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "secret", "val")

	if err := w.Write(ctx, "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "MY_SECRET=val" {
		t.Errorf("expected MY_SECRET=val, got %q", got)
	}
}

func TestWrite_UnknownProvider_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	reg := provider.NewRegistry()
	exp := environ.New(reg)
	w := environ.NewWriter(exp, &buf)

	err := w.Write(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNewWriter_NilWriter_DefaultsToStdout(t *testing.T) {
	reg := provider.NewRegistry()
	exp := environ.New(reg)
	w := environ.NewWriter(exp, nil)
	if w == nil {
		t.Fatal("expected non-nil Writer")
	}
}
