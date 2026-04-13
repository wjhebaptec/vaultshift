package index_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/index"
)

type mockProvider struct {
	keys []string
	err  error
}

func (m *mockProvider) List(_ context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.keys, nil
}

func TestBuild_PopulatesIndex(t *testing.T) {
	idx := index.New()
	providers := map[string]index.Provider{
		"aws": &mockProvider{keys: []string{"prod/db/pass", "prod/api/key"}},
		"gcp": &mockProvider{keys: []string{"staging/db/pass"}},
	}

	errs := index.Build(context.Background(), idx, providers)
	if errs != nil {
		t.Fatalf("unexpected errors: %v", errs)
	}

	all := idx.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
}

func TestBuild_PartialFailure_ReturnsErrors(t *testing.T) {
	idx := index.New()
	providers := map[string]index.Provider{
		"aws": &mockProvider{keys: []string{"prod/db/pass"}},
		"gcp": &mockProvider{err: errors.New("connection refused")},
	}

	errs := index.Build(context.Background(), idx, providers)
	if errs == nil {
		t.Fatal("expected errors map to be non-nil")
	}
	if _, ok := errs["gcp"]; !ok {
		t.Error("expected error for gcp provider")
	}

	// aws keys should still be indexed
	results := idx.SearchPrefix("prod/")
	if len(results) != 1 {
		t.Fatalf("expected 1 aws entry, got %d", len(results))
	}
}

func TestBuild_AllFail_ReturnsAllErrors(t *testing.T) {
	idx := index.New()
	providers := map[string]index.Provider{
		"aws": &mockProvider{err: errors.New("timeout")},
		"vault": &mockProvider{err: errors.New("unauthorized")},
	}

	errs := index.Build(context.Background(), idx, providers)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
	if len(idx.All()) != 0 {
		t.Fatal("expected empty index when all providers fail")
	}
}

func TestBuild_EmptyProviders_ReturnsNil(t *testing.T) {
	idx := index.New()
	errs := index.Build(context.Background(), idx, map[string]index.Provider{})
	if errs != nil {
		t.Fatalf("expected nil errors for empty providers, got %v", errs)
	}
}
