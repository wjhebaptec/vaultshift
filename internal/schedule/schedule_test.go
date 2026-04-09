package schedule_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vaultshift/vaultshift/internal/schedule"
)

func TestRegister_ValidJob(t *testing.T) {
	s := schedule.New()
	err := s.Register(schedule.Job{
		Name:     "rotate-db",
		Interval: 10 * time.Millisecond,
		Task:     func(ctx context.Context) error { return nil },
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	names := s.JobNames()
	if len(names) != 1 || names[0] != "rotate-db" {
		t.Errorf("unexpected job names: %v", names)
	}
}

func TestRegister_DuplicateName(t *testing.T) {
	s := schedule.New()
	job := schedule.Job{Name: "dup", Interval: time.Second, Task: func(ctx context.Context) error { return nil }}
	_ = s.Register(job)
	err := s.Register(job)
	if err == nil {
		t.Fatal("expected error for duplicate job name")
	}
}

func TestRegister_InvalidJob(t *testing.T) {
	s := schedule.New()
	if err := s.Register(schedule.Job{Name: "", Interval: time.Second, Task: func(ctx context.Context) error { return nil }}); err == nil {
		t.Error("expected error for empty name")
	}
	if err := s.Register(schedule.Job{Name: "x", Interval: 0, Task: func(ctx context.Context) error { return nil }}); err == nil {
		t.Error("expected error for zero interval")
	}
	if err := s.Register(schedule.Job{Name: "y", Interval: time.Second, Task: nil}); err == nil {
		t.Error("expected error for nil task")
	}
}

func TestStart_ExecutesTask(t *testing.T) {
	s := schedule.New()
	var count int64
	_ = s.Register(schedule.Job{
		Name:     "counter",
		Interval: 20 * time.Millisecond,
		Task: func(ctx context.Context) error {
			atomic.AddInt64(&count, 1)
			return nil
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	time.Sleep(75 * time.Millisecond)
	s.Stop()
	if c := atomic.LoadInt64(&count); c < 2 {
		t.Errorf("expected task to run at least 2 times, got %d", c)
	}
}

func TestStop_HaltsExecution(t *testing.T) {
	s := schedule.New()
	var count int64
	_ = s.Register(schedule.Job{
		Name:     "stopper",
		Interval: 15 * time.Millisecond,
		Task: func(ctx context.Context) error {
			atomic.AddInt64(&count, 1)
			return nil
		},
	})
	ctx := context.Background()
	s.Start(ctx)
	time.Sleep(40 * time.Millisecond)
	s.Stop()
	before := atomic.LoadInt64(&count)
	time.Sleep(40 * time.Millisecond)
	after := atomic.LoadInt64(&count)
	if after != before {
		t.Errorf("task ran after Stop: before=%d after=%d", before, after)
	}
}

func TestStart_ContextCancellation(t *testing.T) {
	s := schedule.New()
	var count int64
	_ = s.Register(schedule.Job{
		Name:     "ctx-cancel",
		Interval: 15 * time.Millisecond,
		Task: func(ctx context.Context) error {
			atomic.AddInt64(&count, 1)
			return nil
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	s.Start(ctx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)
	before := atomic.LoadInt64(&count)
	time.Sleep(30 * time.Millisecond)
	after := atomic.LoadInt64(&count)
	if after != before {
		t.Errorf("task ran after context cancelled: before=%d after=%d", before, after)
	}
}
