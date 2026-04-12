package scope_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/scope"
)

func setup(t *testing.T, namespace string) (*scope.Scoped, *mock.Provider) {
	t.Helper()
	mp := mock.New()
	s, err := scope.New(mp, namespace, "/")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, mp
}

func TestNew_NilProvider_ReturnsError(t *testing.T) {
	_, err := scope.New(nil, "ns", "/")
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestNew_EmptyNamespace_ReturnsError(t *testing.T) {
	_, err := scope.New(mock.New(), "", "/")
	if err == nil {
		t.Fatal("expected error for empty namespace")
	}
}

func TestPut_AndGet_QualifiesKey(t *testing.T) {
	s, mp := setup(t, "prod")
	ctx := context.Background()

	if err := s.Put(ctx, "db_pass", "secret123"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Verify raw key in underlying provider
	val, err := mp.Get(ctx, "prod/db_pass")
	if err != nil {
		t.Fatalf("raw Get: %v", err)
	}
	if val != "secret123" {
		t.Errorf("expected %q, got %q", "secret123", val)
	}

	// Verify scoped Get strips prefix
	got, err := s.Get(ctx, "db_pass")
	if err != nil {
		t.Fatalf("scoped Get: %v", err)
	}
	if got != "secret123" {
		t.Errorf("expected %q, got %q", "secret123", got)
	}
}

func TestDelete_RemovesQualifiedKey(t *testing.T) {
	s, mp := setup(t, "staging")
	ctx := context.Background()

	_ = mp.Put(ctx, "staging/token", "abc")
	if err := s.Delete(ctx, "token"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := mp.Get(ctx, "staging/token")
	if err == nil {
		t.Fatal("expected key to be deleted")
	}
}

func TestList_ReturnsOnlyScopedKeys(t *testing.T) {
	s, mp := setup(t, "prod")
	ctx := context.Background()

	_ = mp.Put(ctx, "prod/alpha", "1")
	_ = mp.Put(ctx, "prod/beta", "2")
	_ = mp.Put(ctx, "dev/gamma", "3")

	keys, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
	for _, k := range keys {
		if k == "gamma" {
			t.Error("out-of-scope key leaked into List result")
		}
	}
}

func TestInScope_ReturnsCorrectly(t *testing.T) {
	s, _ := setup(t, "prod")
	if !s.InScope("mykey") {
		t.Error("expected in-scope to be true")
	}
}

func TestGet_PropagatesProviderError(t *testing.T) {
	s, _ := setup(t, "prod")
	_, err := s.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if errors.Is(err, scope.ErrOutOfScope) {
		t.Error("should not be an out-of-scope error")
	}
}
