package gradient_test

import (
	"testing"

	"github.com/vaultshift/internal/gradient"
)

func setupShifter(t *testing.T) (*gradient.Gradient, *gradient.Shifter) {
	t.Helper()
	reg := buildRegistry(t, "alpha", "beta")
	g, err := gradient.New(reg, []string{"alpha", "beta"}, []float64{50, 50})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sh, err := gradient.NewShifter(g)
	if err != nil {
		t.Fatalf("NewShifter: %v", err)
	}
	return g, sh
}

func TestNewShifter_NilGradient_ReturnsError(t *testing.T) {
	_, err := gradient.NewShifter(nil)
	if err == nil {
		t.Fatal("expected error for nil gradient")
	}
}

func TestApply_UpdatesWeights(t *testing.T) {
	_, sh := setupShifter(t)
	err := sh.Apply([]gradient.Shift{
		{Name: "alpha", Weight: 90},
		{Name: "beta", Weight: 10},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	weights := sh.Weights()
	if weights["alpha"] != 90 {
		t.Errorf("expected alpha=90, got %v", weights["alpha"])
	}
	if weights["beta"] != 10 {
		t.Errorf("expected beta=10, got %v", weights["beta"])
	}
}

func TestApply_ZeroTotalWeight_ReturnsError(t *testing.T) {
	_, sh := setupShifter(t)
	err := sh.Apply([]gradient.Shift{
		{Name: "alpha", Weight: 0},
		{Name: "beta", Weight: 0},
	})
	if err == nil {
		t.Fatal("expected error for zero total weight")
	}
}

func TestApply_UnknownProvider_ReturnsError(t *testing.T) {
	_, sh := setupShifter(t)
	err := sh.Apply([]gradient.Shift{
		{Name: "alpha", Weight: 50},
		{Name: "unknown", Weight: 50},
	})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestApply_WrongCount_ReturnsError(t *testing.T) {
	_, sh := setupShifter(t)
	err := sh.Apply([]gradient.Shift{
		{Name: "alpha", Weight: 100},
	})
	if err == nil {
		t.Fatal("expected error for wrong shift count")
	}
}

func TestWeights_ReturnsCurrentDistribution(t *testing.T) {
	_, sh := setupShifter(t)
	w := sh.Weights()
	if len(w) != 2 {
		t.Fatalf("expected 2 weights, got %d", len(w))
	}
	if w["alpha"] != 50 || w["beta"] != 50 {
		t.Errorf("unexpected initial weights: %v", w)
	}
}
