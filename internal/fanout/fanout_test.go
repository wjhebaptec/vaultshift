package fanout_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/fanout"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupFanout(t *testing.T, names ...string) (*fanout.Fanout, map[string]*mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	providers := make(map[string]*mock.Provider)
	for _, name := range names {
		mp := mock.New()
		providers[name] = mp
		reg.Register(name, mp)
	}
	f, err := fanout.New(reg, names)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return f, providers
}

func TestPut_AllSucceed(t *testing.T) {
	f, providers := setupFanout(t, "aws", "gcp", "vault")
	results := f.Put(context.Background(), "db/pass", "s3cr3t")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("provider %s: unexpected error: %v", r.Provider, r.Err)
		}
	}
	for name, mp := range providers {
		v, err := mp.Get(context.Background(), "db/pass")
		if err != nil || v != "s3cr3t" {
			t.Errorf("provider %s: expected value s3cr3t, got %q err %v", name, v, err)
		}
	}
}

func TestPut_UnknownProvider_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	mp := mock.New()
	reg.Register("aws", mp)

	f, err := fanout.New(reg, []string{"aws", "missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := f.Put(context.Background(), "key", "val")
	var found bool
	for _, r := range results {
		if r.Provider == "missing" && r.Err != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected error result for unknown provider 'missing'")
	}
}

func TestHasFailures_True(t *testing.T) {
	results := []fanout.Result{
		{Provider: "aws", Key: "k", Err: nil},
		{Provider: "gcp", Key: "k", Err: errors.New("boom")},
	}
	if !fanout.HasFailures(results) {
		t.Error("expected HasFailures to return true")
	}
}

func TestHasFailures_False(t *testing.T) {
	results := []fanout.Result{
		{Provider: "aws", Key: "k", Err: nil},
		{Provider: "gcp", Key: "k", Err: nil},
	}
	if fanout.HasFailures(results) {
		t.Error("expected HasFailures to return false")
	}
}

func TestNew_NilRegistry_ReturnsError(t *testing.T) {
	_, err := fanout.New(nil, []string{"aws"})
	if err == nil {
		t.Error("expected error for nil registry")
	}
}

func TestNew_NoTargets_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	_, err := fanout.New(reg, nil)
	if err == nil {
		t.Error("expected error for empty targets")
	}
}
