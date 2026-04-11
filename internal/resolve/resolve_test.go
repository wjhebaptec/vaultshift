package resolve_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/resolve"
)

// stubProvider returns values from a fixed map; returns error for missing keys.
type stubProvider struct {
	data map[string]string
}

func (s *stubProvider) Get(_ context.Context, key string) (string, error) {
	if v, ok := s.data[key]; ok {
		return v, nil
	}
	return "", errors.New("not found")
}

func TestResolve_DirectKey(t *testing.T) {
	p := &stubProvider{data: map[string]string{"db/pass": "secret123"}}
	r := resolve.New([]resolve.Provider{p})
	v, err := r.Resolve(context.Background(), "db/pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "secret123" {
		t.Errorf("expected secret123, got %s", v)
	}
}

func TestResolve_AliasedKey(t *testing.T) {
	p := &stubProvider{data: map[string]string{"prod/db/password": "aliased"}}
	r := resolve.New([]resolve.Provider{p}, resolve.WithAlias("db_pass", "prod/db/password"))
	v, err := r.Resolve(context.Background(), "db_pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "aliased" {
		t.Errorf("expected aliased, got %s", v)
	}
}

func TestResolve_FallbackChain(t *testing.T) {
	p1 := &stubProvider{data: map[string]string{}}
	p2 := &stubProvider{data: map[string]string{"key": "fromSecond"}}
	r := resolve.New([]resolve.Provider{p1, p2})
	v, err := r.Resolve(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "fromSecond" {
		t.Errorf("expected fromSecond, got %s", v)
	}
}

func TestResolve_NotFound(t *testing.T) {
	p := &stubProvider{data: map[string]string{}}
	r := resolve.New([]resolve.Provider{p})
	_, err := r.Resolve(context.Background(), "missing")
	if !errors.Is(err, resolve.ErrNotResolved) {
		t.Errorf("expected ErrNotResolved, got %v", err)
	}
}

func TestResolveAll_PartialErrors(t *testing.T) {
	p := &stubProvider{data: map[string]string{"a": "1", "b": "2"}}
	r := resolve.New([]resolve.Provider{p})
	out, errs := r.ResolveAll(context.Background(), []string{"a", "b", "c"})
	if len(out) != 2 {
		t.Errorf("expected 2 resolved, got %d", len(out))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	if out["a"] != "1" || out["b"] != "2" {
		t.Error("unexpected resolved values")
	}
}

func TestResolveAll_AllResolved(t *testing.T) {
	p := &stubProvider{data: map[string]string{"x": "10", "y": "20"}}
	r := resolve.New([]resolve.Provider{p})
	out, errs := r.ResolveAll(context.Background(), []string{"x", "y"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}
	if len(out) != 2 {
		t.Errorf("expected 2 results, got %d", len(out))
	}
}
