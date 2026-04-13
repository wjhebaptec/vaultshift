package coalesce_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/coalesce"
)

// stubProvider is a simple in-memory provider for testing.
type stubProvider struct {
	data map[string]string
	err  error
}

func (s *stubProvider) Get(_ context.Context, key string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	v, ok := s.data[key]
	if !ok {
		return "", nil
	}
	return v, nil
}

func TestGet_ReturnsFirstNonEmpty(t *testing.T) {
	c := coalesce.New()
	c.Add("empty", &stubProvider{data: map[string]string{}})
	c.Add("second", &stubProvider{data: map[string]string{"db_pass": "secret"}})

	v, src, err := c.Get(context.Background(), "db_pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "secret" {
		t.Errorf("expected 'secret', got %q", v)
	}
	if src != "second" {
		t.Errorf("expected source 'second', got %q", src)
	}
}

func TestGet_NotFoundInAny(t *testing.T) {
	c := coalesce.New()
	c.Add("a", &stubProvider{data: map[string]string{}})
	c.Add("b", &stubProvider{data: map[string]string{}})

	_, _, err := c.Get(context.Background(), "missing")
	if !errors.Is(err, coalesce.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_SkipsProviderErrors(t *testing.T) {
	c := coalesce.New()
	c.Add("broken", &stubProvider{err: errors.New("unavailable")})
	c.Add("ok", &stubProvider{data: map[string]string{"key": "val"}})

	v, src, err := c.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "val" || src != "ok" {
		t.Errorf("expected val/ok, got %q/%q", v, src)
	}
}

func TestAdd_NilProvider_ReturnsError(t *testing.T) {
	c := coalesce.New()
	if err := c.Add("x", nil); err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestAdd_EmptyName_ReturnsError(t *testing.T) {
	c := coalesce.New()
	if err := c.Add("", &stubProvider{}); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestGetAll_ReturnsAllNonEmpty(t *testing.T) {
	c := coalesce.New()
	c.Add("p1", &stubProvider{data: map[string]string{"k": "v1"}})
	c.Add("p2", &stubProvider{data: map[string]string{"k": "v2"}})
	c.Add("p3", &stubProvider{data: map[string]string{}})

	all := c.GetAll(context.Background(), "k")
	if len(all) != 2 {
		t.Fatalf("expected 2 results, got %d", len(all))
	}
	if all["p1"] != "v1" || all["p2"] != "v2" {
		t.Errorf("unexpected values: %v", all)
	}
}

func TestGet_NoProviders_ReturnsNotFound(t *testing.T) {
	c := coalesce.New()
	_, _, err := c.Get(context.Background(), "any")
	if !errors.Is(err, coalesce.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
