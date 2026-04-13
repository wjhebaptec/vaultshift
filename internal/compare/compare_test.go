package compare_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/compare"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupComparer(t *testing.T) (*compare.Comparer, *provider.Registry) {
	t.Helper()
	reg := provider.NewRegistry()

	a := mock.New()
	b := mock.New()
	reg.Register("providerA", a)
	reg.Register("providerB", b)

	c, err := compare.New(reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return c, reg
}

func TestCompare_AllMatch(t *testing.T) {
	c, reg := setupComparer(t)
	ctx := context.Background()

	for _, name := range []string{"providerA", "providerB"} {
		p, _ := reg.Get(name)
		_ = p.PutSecret(ctx, "db/pass", "s3cr3t")
	}

	res, err := c.Compare(ctx, "db/pass", []string{"providerA", "providerB"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Match {
		t.Errorf("expected Match=true, got false; values=%v", res.Values)
	}
}

func TestCompare_Mismatch(t *testing.T) {
	c, reg := setupComparer(t)
	ctx := context.Background()

	pa, _ := reg.Get("providerA")
	pb, _ := reg.Get("providerB")
	_ = pa.PutSecret(ctx, "db/pass", "alpha")
	_ = pb.PutSecret(ctx, "db/pass", "beta")

	res, err := c.Compare(ctx, "db/pass", []string{"providerA", "providerB"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Match {
		t.Error("expected Match=false, got true")
	}
}

func TestCompare_MissingInOneProvider(t *testing.T) {
	c, reg := setupComparer(t)
	ctx := context.Background()

	pa, _ := reg.Get("providerA")
	_ = pa.PutSecret(ctx, "db/pass", "s3cr3t")
	// providerB has no entry

	res, err := c.Compare(ctx, "db/pass", []string{"providerA", "providerB"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Match {
		t.Error("expected Match=false when key is missing in one provider")
	}
	if len(res.Missing) != 1 || res.Missing[0] != "providerB" {
		t.Errorf("expected Missing=[providerB], got %v", res.Missing)
	}
}

func TestCompare_EmptyKey_ReturnsError(t *testing.T) {
	c, _ := setupComparer(t)
	_, err := c.Compare(context.Background(), "", []string{"providerA"})
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestCompare_UnknownProvider_ReturnsError(t *testing.T) {
	c, _ := setupComparer(t)
	_, err := c.Compare(context.Background(), "key", []string{"ghost"})
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestCompareAll_ReturnsAllResults(t *testing.T) {
	c, reg := setupComparer(t)
	ctx := context.Background()

	pa, _ := reg.Get("providerA")
	pb, _ := reg.Get("providerB")
	for _, k := range []string{"k1", "k2"} {
		_ = pa.PutSecret(ctx, k, "val")
		_ = pb.PutSecret(ctx, k, "val")
	}

	results, err := c.CompareAll(ctx, []string{"k1", "k2"}, []string{"providerA", "providerB"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Match {
			t.Errorf("expected Match=true for key %q", r.Key)
		}
	}
}

func TestNew_NilRegistry_ReturnsError(t *testing.T) {
	_, err := compare.New(nil)
	if err == nil {
		t.Error("expected error for nil registry")
	}
}
