package partition_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/partition"
	"github.com/vaultshift/internal/provider/mock"
)

func setup(t *testing.T) (*partition.Partitioner, map[string]*mock.Provider) {
	t.Helper()
	provA := mock.New()
	provB := mock.New()
	router := func(key string) string {
		if len(key) > 0 && key[0] == 'b' {
			return "b"
		}
		return "a"
	}
	p, err := partition.New(router, map[string]interface{ Put(context.Context, string, string) error }{
	})
	_ = p
	_ = err
	// rebuild with correct type
	prov := map[string]interface{}{}
	_ = prov
	part, err := partition.New(router, map[string]interface {
		Put(ctx context.Context, key, value string) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
	}{
		"a": provA,
		"b": provB,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return part, map[string]*mock.Provider{"a": provA, "b": provB}
}

func TestNew_NilRouter_ReturnsError(t *testing.T) {
	_, err := partition.New(nil, map[string]interface {
		Put(ctx context.Context, key, value string) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
	}{"a": mock.New()})
	if err == nil {
		t.Fatal("expected error for nil router")
	}
}

func TestNew_NoProviders_ReturnsError(t *testing.T) {
	_, err := partition.New(func(k string) string { return "a" }, map[string]interface {
		Put(ctx context.Context, key, value string) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
	}{})
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
}

func TestPut_AndGet_RoutedCorrectly(t *testing.T) {
	ctx := context.Background()
	provA := mock.New()
	provB := mock.New()
	router := func(key string) string {
		if len(key) > 0 && key[0] == 'b' {
			return "b"
		}
		return "a"
	}
	p, _ := partition.New(router, map[string]interface {
		Put(ctx context.Context, key, value string) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
	}{"a": provA, "b": provB})

	if err := p.Put(ctx, "alpha", "val-a"); err != nil {
		t.Fatalf("Put alpha: %v", err)
	}
	if err := p.Put(ctx, "beta", "val-b"); err != nil {
		t.Fatalf("Put beta: %v", err)
	}

	v, err := p.Get(ctx, "alpha")
	if err != nil || v != "val-a" {
		t.Fatalf("Get alpha: got %q, err %v", v, err)
	}
	v, err = p.Get(ctx, "beta")
	if err != nil || v != "val-b" {
		t.Fatalf("Get beta: got %q, err %v", v, err)
	}

	// Ensure cross-routing isolation
	if _, err := provA.Get(ctx, "beta"); err == nil {
		t.Fatal("beta should not exist in provider a")
	}
}

func TestGet_UnknownProvider_ReturnsError(t *testing.T) {
	ctx := context.Background()
	p, _ := partition.New(func(k string) string { return "missing" }, map[string]interface {
		Put(ctx context.Context, key, value string) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
	}{"a": mock.New()})
	if _, err := p.Get(ctx, "any"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestDelete_RemovesFromCorrectProvider(t *testing.T) {
	ctx := context.Background()
	provA := mock.New()
	p, _ := partition.New(func(k string) string { return "a" }, map[string]interface {
		Put(ctx context.Context, key, value string) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
	}{"a": provA})

	_ = p.Put(ctx, "key1", "v")
	if err := p.Delete(ctx, "key1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := provA.Get(ctx, "key1"); err == nil {
		t.Fatal("expected key1 to be deleted")
	}
}
