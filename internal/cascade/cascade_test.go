package cascade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/cascade"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupCascade(t *testing.T, primary string, chain []string, opts ...cascade.Option) (*cascade.Cascade, map[string]*mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	providers := make(map[string]*mock.Provider)
	for _, name := range append([]string{primary}, chain...) {
		m := mock.New()
		providers[name] = m
		reg.Register(name, m)
	}
	c, err := cascade.New(primary, chain, reg, opts...)
	if err != nil {
		t.Fatalf("cascade.New: %v", err)
	}
	return c, providers
}

func TestPut_AllSucceed(t *testing.T) {
	c, providers := setupCascade(t, "primary", []string{"secondary", "tertiary"})
	results, err := c.Put(context.Background(), "api/key", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if cascade.HasFailures(results) {
		t.Error("expected no failures")
	}
	for name, m := range providers {
		v, _ := m.Get(context.Background(), "api/key")
		if v != "secret" {
			t.Errorf("provider %q: expected 'secret', got %q", name, v)
		}
	}
}

func TestPut_SecondaryFails_ContinuesByDefault(t *testing.T) {
	reg := provider.NewRegistry()
	prim := mock.New()
	fail := mock.New()
	fail.ForceError(errors.New("write error"))
	good := mock.New()
	reg.Register("primary", prim)
	reg.Register("fail", fail)
	reg.Register("good", good)
	c, err := cascade.New("primary", []string{"fail", "good"}, reg)
	if err != nil {
		t.Fatalf("cascade.New: %v", err)
	}
	results, _ := c.Put(context.Background(), "k", "v")
	if !cascade.HasFailures(results) {
		t.Error("expected a failure in results")
	}
	// good provider should still have been written
	v, _ := good.Get(context.Background(), "k")
	if v != "v" {
		t.Errorf("good provider should have received write, got %q", v)
	}
}

func TestPut_StopOnFailure_HaltsChain(t *testing.T) {
	reg := provider.NewRegistry()
	prim := mock.New()
	fail := mock.New()
	fail.ForceError(errors.New("write error"))
	skipped := mock.New()
	reg.Register("primary", prim)
	reg.Register("fail", fail)
	reg.Register("skipped", skipped)
	c, err := cascade.New("primary", []string{"fail", "skipped"}, reg, cascade.WithStopOnFailure())
	if err != nil {
		t.Fatalf("cascade.New: %v", err)
	}
	results, _ := c.Put(context.Background(), "k", "v")
	if len(results) != 2 {
		t.Fatalf("expected 2 results (primary+fail), got %d", len(results))
	}
	v, _ := skipped.Get(context.Background(), "k")
	if v != "" {
		t.Error("skipped provider should not have received a write")
	}
}

func TestNew_EmptyPrimary_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	_, err := cascade.New("", []string{"secondary"}, reg)
	if err == nil {
		t.Fatal("expected error for empty primary")
	}
}

func TestNew_EmptyChain_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	_, err := cascade.New("primary", nil, reg)
	if !errors.Is(err, cascade.ErrNoProviders) {
		t.Fatalf("expected ErrNoProviders, got %v", err)
	}
}
