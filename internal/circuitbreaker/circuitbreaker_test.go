package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/circuitbreaker"
)

func newCB(maxFailures int, resetTimeout time.Duration) *circuitbreaker.CircuitBreaker {
	return circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:  maxFailures,
		ResetTimeout: resetTimeout,
	})
}

func TestAllow_InitiallyClosed(t *testing.T) {
	cb := newCB(3, time.Second)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordFailure_OpensCircuit(t *testing.T) {
	cb := newCB(3, time.Second)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatal("expected circuit to be open")
	}
	if err := cb.Allow(); err == nil {
		t.Fatal("expected error when circuit is open")
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	cb := newCB(2, time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatal("expected open state")
	}
	cb.RecordSuccess()
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatal("expected closed state after success")
	}
	if cb.Failures() != 0 {
		t.Fatalf("expected 0 failures, got %d", cb.Failures())
	}
}

func TestAllow_HalfOpenAfterTimeout(t *testing.T) {
	cb := newCB(1, 50*time.Millisecond)
	cb.RecordFailure()
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatal("expected open")
	}
	time.Sleep(60 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil after timeout, got %v", err)
	}
	if cb.State() != circuitbreaker.StateHalfOpen {
		t.Fatal("expected half-open state")
	}
}

func TestDefaultConfig_UsedOnZeroValues(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{})
	def := circuitbreaker.DefaultConfig()
	for i := 0; i < def.MaxFailures; i++ {
		cb.RecordFailure()
	}
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatal("expected circuit to open at default threshold")
	}
}

func TestFailures_TracksCount(t *testing.T) {
	cb := newCB(10, time.Second)
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
	}
	if cb.Failures() != 4 {
		t.Fatalf("expected 4 failures, got %d", cb.Failures())
	}
}
