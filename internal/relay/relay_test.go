package relay_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/relay"
)

func setupRelay(t *testing.T) (*relay.Relay, *mock.Provider, *mock.Provider) {
	t.Helper()
	src := mock.New()
	dst := mock.New()
	reg := provider.NewRegistry()
	reg.Register("src", src)
	reg.Register("dst", dst)
	return relay.New(reg), src, dst
}

func TestForward_CopiesSecret(t *testing.T) {
	rl, src, dst := setupRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "api/key", "secret123")
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "api/key", DestProvider: "dst"})
	results := rl.Forward(ctx)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("expected success, got %v", results[0].Err)
	}
	v, err := dst.Get(ctx, "api/key")
	if err != nil || v != "secret123" {
		t.Fatalf("expected secret123, got %q %v", v, err)
	}
}

func TestForward_WithKeyRename(t *testing.T) {
	rl, src, dst := setupRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "old/key", "value")
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "old/key", DestProvider: "dst", DestKey: "new/key"})
	rl.Forward(ctx)
	v, err := dst.Get(ctx, "new/key")
	if err != nil || v != "value" {
		t.Fatalf("expected value at new/key, got %q %v", v, err)
	}
}

func TestForward_WithTransform(t *testing.T) {
	rl, src, dst := setupRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "k", "hello")
	_ = rl.Register(relay.Rule{
		SourceProvider: "src", SourceKey: "k",
		DestProvider: "dst",
		Transform: func(v string) (string, error) { return strings.ToUpper(v), nil },
	})
	rl.Forward(ctx)
	v, _ := dst.Get(ctx, "k")
	if v != "HELLO" {
		t.Fatalf("expected HELLO, got %q", v)
	}
}

func TestForward_UnknownSourceProvider(t *testing.T) {
	rl, _, _ := setupRelay(t)
	_ = rl.Register(relay.Rule{SourceProvider: "nope", SourceKey: "k", DestProvider: "dst"})
	results := rl.Forward(context.Background())
	if results[0].Err == nil {
		t.Fatal("expected error for unknown source provider")
	}
}

func TestForward_UnknownDestProvider(t *testing.T) {
	rl, src, _ := setupRelay(t)
	ctx := context.Background()
	_ = src.Put(ctx, "k", "v")
	_ = rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "k", DestProvider: "nope"})
	results := rl.Forward(ctx)
	if results[0].Err == nil {
		t.Fatal("expected error for unknown dest provider")
	}
}

func TestRegister_MissingFields(t *testing.T) {
	rl, _, _ := setupRelay(t)
	if err := rl.Register(relay.Rule{SourceKey: "k", DestProvider: "dst"}); err == nil {
		t.Fatal("expected error for missing source provider")
	}
	if err := rl.Register(relay.Rule{SourceProvider: "src", SourceKey: "k"}); err == nil {
		t.Fatal("expected error for missing dest provider")
	}
}

func TestHasFailures(t *testing.T) {
	results := []relay.Result{{}, {Err: fmt.Errorf("oops")}}
	if !relay.HasFailures(results) {
		t.Fatal("expected HasFailures to return true")
	}
	if relay.HasFailures([]relay.Result{{}}) {
		t.Fatal("expected HasFailures to return false")
	}
}
