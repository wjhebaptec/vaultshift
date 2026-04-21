package gradient_test

import (
	"testing"

	"github.com/vaultshift/internal/gradient"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func buildRegistry(t *testing.T, names ...string) *provider.Registry {
	t.Helper()
	reg := provider.NewRegistry()
	for _, n := range names {
		reg.Register(n, mock.New())
	}
	return reg
}

func TestNew_ValidWeights(t *testing.T) {
	reg := buildRegistry(t, "a", "b")
	g, err := gradient.New(reg, []string{"a", "b"}, []float64{70, 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil {
		t.Fatal("expected non-nil gradient")
	}
}

func TestNew_NoProviders_ReturnsError(t *testing.T) {
	reg := buildRegistry(t)
	_, err := gradient.New(reg, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
}

func TestNew_MismatchedLengths_ReturnsError(t *testing.T) {
	reg := buildRegistry(t, "a")
	_, err := gradient.New(reg, []string{"a"}, []float64{50, 50})
	if err == nil {
		t.Fatal("expected error for mismatched lengths")
	}
}

func TestNew_NegativeWeight_ReturnsError(t *testing.T) {
	reg := buildRegistry(t, "a", "b")
	_, err := gradient.New(reg, []string{"a", "b"}, []float64{-1, 50})
	if err == nil {
		t.Fatal("expected error for negative weight")
	}
}

func TestNew_ZeroTotalWeight_ReturnsError(t *testing.T) {
	reg := buildRegistry(t, "a")
	_, err := gradient.New(reg, []string{"a"}, []float64{0})
	if err == nil {
		t.Fatal("expected error for zero total weight")
	}
}

func TestNew_UnknownProvider_ReturnsError(t *testing.T) {
	reg := buildRegistry(t)
	_, err := gradient.New(reg, []string{"missing"}, []float64{100})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestGet_RoutesToProvider(t *testing.T) {
	m := mock.New()
	_ = m.Put("token", "secret123")
	reg := provider.NewRegistry()
	reg.Register("only", m)

	g, err := gradient.New(reg, []string{"only"}, []float64{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := g.Get("token")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "secret123" {
		t.Errorf("expected secret123, got %s", val)
	}
}

func TestPut_RoutesToProvider(t *testing.T) {
	m := mock.New()
	reg := provider.NewRegistry()
	reg.Register("only", m)

	g, err := gradient.New(reg, []string{"only"}, []float64{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := g.Put("k", "v"); err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	val, _ := m.Get("k")
	if val != "v" {
		t.Errorf("expected v, got %s", val)
	}
}
