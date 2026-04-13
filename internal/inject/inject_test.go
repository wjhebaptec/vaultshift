package inject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/inject"
)

// stubProvider implements inject.Provider for testing.
type stubProvider struct {
	data map[string]string
}

func (s *stubProvider) Get(_ context.Context, key string) (string, error) {
	v, ok := s.data[key]
	if !ok {
		return "", errors.New("not found: " + key)
	}
	return v, nil
}

func TestNew_NilProvider_ReturnsError(t *testing.T) {
	_, err := inject.New(nil)
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestInject_AllKeysResolved(t *testing.T) {
	p := &stubProvider{data: map[string]string{"db/pass": "secret", "api/key": "abc"}}
	inj, _ := inject.New(p)
	target := inject.MapTarget{}
	errs := inj.Inject(context.Background(), []string{"db/pass", "api/key"}, target)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if target["db/pass"] != "secret" || target["api/key"] != "abc" {
		t.Fatalf("unexpected target contents: %v", map[string]string(target))
	}
}

func TestInject_MissingKey_ReturnsError(t *testing.T) {
	p := &stubProvider{data: map[string]string{}}
	inj, _ := inject.New(p)
	target := inject.MapTarget{}
	errs := inj.Inject(context.Background(), []string{"missing"}, target)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestInject_PartialFailure_ContinuesRemaining(t *testing.T) {
	p := &stubProvider{data: map[string]string{"good": "val"}}
	inj, _ := inject.New(p)
	target := inject.MapTarget{}
	errs := inj.Inject(context.Background(), []string{"missing", "good"}, target)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if target["good"] != "val" {
		t.Fatal("expected 'good' key to be injected")
	}
}

func TestInject_WithPrefix_StripsPrefix(t *testing.T) {
	p := &stubProvider{data: map[string]string{"prod/db/pass": "s3cr3t"}}
	inj, _ := inject.New(p, inject.WithPrefix("prod/"))
	target := inject.MapTarget{}
	inj.Inject(context.Background(), []string{"prod/db/pass"}, target)
	if target["db/pass"] != "s3cr3t" {
		t.Fatalf("expected stripped key, got: %v", map[string]string(target))
	}
}

func TestInjectMap_ReturnsMap(t *testing.T) {
	p := &stubProvider{data: map[string]string{"x": "1", "y": "2"}}
	inj, _ := inject.New(p)
	m, errs := inj.InjectMap(context.Background(), []string{"x", "y"})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if m["x"] != "1" || m["y"] != "2" {
		t.Fatalf("unexpected map: %v", m)
	}
}
