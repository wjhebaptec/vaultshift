package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultshift/internal/circuitbreaker"
)

type stubProvider struct {
	getErr error
	putErr error
	val    string
}

func (s *stubProvider) Get(_ context.Context, _ string) (string, error) {
	return s.val, s.getErr
}
func (s *stubProvider) Put(_ context.Context, _, _ string) error { return s.putErr }

func TestWrapGet_SuccessResetsBreakerFailures(t *testing.T) {
	cb := newCB(5, time.Second)
	cb.RecordFailure()
	stub := &stubProvider{val: "secret"}
	wp := circuitbreaker.Wrap("test", stub, cb)
	val, err := wp.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret" {
		t.Fatalf("expected 'secret', got %q", val)
	}
	if cb.Failures() != 0 {
		t.Fatalf("expected 0 failures after success, got %d", cb.Failures())
	}
}

func TestWrapGet_FailureIncrementsBreaker(t *testing.T) {
	cb := newCB(5, time.Second)
	stub := &stubProvider{getErr: errors.New("timeout")}
	wp := circuitbreaker.Wrap("test", stub, cb)
	_, err := wp.Get(context.Background(), "key")
	if err == nil {
		t.Fatal("expected error")
	}
	if cb.Failures() != 1 {
		t.Fatalf("expected 1 failure, got %d", cb.Failures())
	}
}

func TestWrapGet_BlockedWhenOpen(t *testing.T) {
	cb := newCB(1, time.Minute)
	cb.RecordFailure()
	stub := &stubProvider{val: "secret"}
	wp := circuitbreaker.Wrap("aws", stub, cb)
	_, err := wp.Get(context.Background(), "key")
	if err == nil {
		t.Fatal("expected circuit open error")
	}
	if !errors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestWrapPut_FailureOpensCircuit(t *testing.T) {
	cb := newCB(2, time.Second)
	stub := &stubProvider{putErr: errors.New("write error")}
	wp := circuitbreaker.Wrap("gcp", stub, cb)
	for i := 0; i < 2; i++ {
		_ = wp.Put(context.Background(), "k", "v")
	}
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatal("expected circuit to open after failures")
	}
}

func TestBreaker_ReturnsUnderlyingCB(t *testing.T) {
	cb := newCB(3, time.Second)
	stub := &stubProvider{}
	wp := circuitbreaker.Wrap("vault", stub, cb)
	if wp.Breaker() != cb {
		t.Fatal("expected same circuit breaker instance")
	}
}
