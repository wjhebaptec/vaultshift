package drain_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vaultshift/internal/drain"
)

func TestAcquire_SuccessWhenOpen(t *testing.T) {
	d := drain.New()
	if err := d.Acquire(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	d.Release()
}

func TestAcquire_FailsAfterDrain(t *testing.T) {
	d := drain.New(drain.WithTimeout(100 * time.Millisecond))
	go func() { _ = d.Drain(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	if err := d.Acquire(); err == nil {
		t.Fatal("expected error after drain, got nil")
	}
}

func TestDrain_WaitsForInFlight(t *testing.T) {
	d := drain.New(drain.WithTimeout(2 * time.Second))
	var started, released sync.WaitGroup
	started.Add(1)
	released.Add(1)

	_ = d.Acquire()
	go func() {
		started.Done()
		time.Sleep(50 * time.Millisecond)
		d.Release()
		released.Done()
	}()

	started.Wait()
	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("drain returned error: %v", err)
	}
	released.Wait()
}

func TestDrain_TimesOut(t *testing.T) {
	d := drain.New(drain.WithTimeout(50 * time.Millisecond))
	_ = d.Acquire() // never released

	err := d.Drain(context.Background())
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestIsClosed_FalseInitially(t *testing.T) {
	d := drain.New()
	if d.IsClosed() {
		t.Fatal("expected drainer to be open initially")
	}
}

func TestIsClosed_TrueAfterDrain(t *testing.T) {
	d := drain.New(drain.WithTimeout(100 * time.Millisecond))
	go func() { _ = d.Drain(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	if !d.IsClosed() {
		t.Fatal("expected drainer to be closed after Drain called")
	}
}
