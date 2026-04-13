package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/health"
)

type mockChecker struct {
	name string
	err  error
}

func (m *mockChecker) Name() string { return m.name }
func (m *mockChecker) Ping(_ context.Context) error { return m.err }

func TestRegister_NilChecker(t *testing.T) {
	mon := health.New()
	if err := mon.Register(nil); err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestCheckAll_AllHealthy(t *testing.T) {
	mon := health.New()
	_ = mon.Register(&mockChecker{name: "aws"})
	_ = mon.Register(&mockChecker{name: "gcp"})

	results := mon.CheckAll(context.Background())
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != health.StatusOK {
			t.Errorf("expected OK for %s, got %s", r.Provider, r.Status)
		}
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Provider, r.Err)
		}
	}
}

func TestCheckAll_DegradedProvider(t *testing.T) {
	mon := health.New()
	_ = mon.Register(&mockChecker{name: "vault", err: errors.New("connection refused")})

	results := mon.CheckAll(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != health.StatusDegraded {
		t.Errorf("expected Degraded, got %s", results[0].Status)
	}
	if results[0].Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestCheck_SingleProvider(t *testing.T) {
	mon := health.New()
	_ = mon.Register(&mockChecker{name: "aws"})

	r, err := mon.Check(context.Background(), "aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Provider != "aws" {
		t.Errorf("expected provider aws, got %s", r.Provider)
	}
	if r.Status != health.StatusOK {
		t.Errorf("expected OK, got %s", r.Status)
	}
}

func TestCheck_UnknownProvider(t *testing.T) {
	mon := health.New()
	_, err := mon.Check(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestCheckAll_LatencyRecorded(t *testing.T) {
	mon := health.New()
	_ = mon.Register(&mockChecker{name: "gcp"})

	results := mon.CheckAll(context.Background())
	if results[0].Latency < 0 {
		t.Error("latency should be non-negative")
	}
	if results[0].CheckedAt.IsZero() {
		t.Error("CheckedAt should be set")
	}
}

// TestRegister_DuplicateName verifies that registering two checkers with the
// same name returns an error on the second registration.
func TestRegister_DuplicateName(t *testing.T) {
	mon := health.New()
	if err := mon.Register(&mockChecker{name: "aws"}); err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}
	if err := mon.Register(&mockChecker{name: "aws"}); err == nil {
		t.Fatal("expected error when registering duplicate provider name")
	}
}
