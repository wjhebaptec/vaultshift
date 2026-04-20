package shred_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/shred"
)

// fakeProvider records calls for assertions.
type fakeProvider struct {
	data    map[string]string
	putErr  error
	delErr  error
	putKeys []string
	delKeys []string
}

func newFake() *fakeProvider {
	return &fakeProvider{data: make(map[string]string)}
}

func (f *fakeProvider) Put(_ context.Context, key, value string) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.data[key] = value
	f.putKeys = append(f.putKeys, key)
	return nil
}

func (f *fakeProvider) Delete(_ context.Context, key string) error {
	if f.delErr != nil {
		return f.delErr
	}
	delete(f.data, key)
	f.delKeys = append(f.delKeys, key)
	return nil
}

func setup(t *testing.T) (*shred.Shredder, *fakeProvider) {
	t.Helper()
	p := newFake()
	p.data["mykey"] = "original"
	s, err := shred.New(map[string]shred.Provider{"mock": p})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, p
}

func TestShred_OverwritesThenDeletes(t *testing.T) {
	s, p := setup(t)
	if err := s.Shred(context.Background(), "mock", "mykey"); err != nil {
		t.Fatalf("Shred: %v", err)
	}
	if _, exists := p.data["mykey"]; exists {
		t.Error("expected key to be deleted")
	}
	if len(p.putKeys) < 1 {
		t.Error("expected at least one overwrite Put call")
	}
}

func TestShred_MultiplePasses(t *testing.T) {
	p := newFake()
	p.data["k"] = "v"
	s, err := shred.New(map[string]shred.Provider{"mock": p}, shred.WithPasses(3))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := s.Shred(context.Background(), "mock", "k"); err != nil {
		t.Fatalf("Shred: %v", err)
	}
	if len(p.putKeys) != 3 {
		t.Errorf("expected 3 overwrite passes, got %d", len(p.putKeys))
	}
}

func TestShred_UnknownProvider(t *testing.T) {
	s, _ := setup(t)
	err := s.Shred(context.Background(), "nonexistent", "k")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestShred_PutError_ReturnsError(t *testing.T) {
	p := newFake()
	p.putErr = errors.New("write denied")
	s, err := shred.New(map[string]shred.Provider{"mock": p})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := s.Shred(context.Background(), "mock", "k"); err == nil {
		t.Fatal("expected error from Put failure")
	}
}

func TestShredAll_AllProviders(t *testing.T) {
	p1, p2 := newFake(), newFake()
	p1.data["secret"] = "a"
	p2.data["secret"] = "b"
	s, err := shred.New(map[string]shred.Provider{"a": p1, "b": p2})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := s.ShredAll(context.Background(), "secret")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if shred.HasFailures(results) {
		for _, r := range results {
			if r.Err != nil {
				t.Errorf("provider %q: %v", r.Provider, r.Err)
			}
		}
	}
}

func TestNew_NoProviders_ReturnsError(t *testing.T) {
	_, err := shred.New(nil)
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
}

func TestHasFailures_False(t *testing.T) {
	results := []shred.Result{{Key: "k", Provider: "p", Err: nil}}
	if shred.HasFailures(results) {
		t.Error("expected no failures")
	}
}
