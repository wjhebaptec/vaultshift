// Package schedule provides rotation scheduling functionality for vaultshift.
// It allows secrets to be rotated on a configurable interval or cron expression.
package schedule

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Job represents a scheduled rotation task.
type Job struct {
	Name     string
	Interval time.Duration
	Task     func(ctx context.Context) error
}

// Scheduler manages and runs periodic rotation jobs.
type Scheduler struct {
	mu   sync.Mutex
	jobs []*jobEntry
}

type jobEntry struct {
	job    Job
	stop   chan struct{}
	running bool
}

// New creates a new Scheduler.
func New() *Scheduler {
	return &Scheduler{}
}

// Register adds a job to the scheduler. Returns an error if a job with the same name exists.
func (s *Scheduler) Register(job Job) error {
	if job.Name == "" {
		return fmt.Errorf("schedule: job name must not be empty")
	}
	if job.Interval <= 0 {
		return fmt.Errorf("schedule: job %q interval must be positive", job.Name)
	}
	if job.Task == nil {
		return fmt.Errorf("schedule: job %q task must not be nil", job.Name)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.jobs {
		if e.job.Name == job.Name {
			return fmt.Errorf("schedule: job %q already registered", job.Name)
		}
	}
	s.jobs = append(s.jobs, &jobEntry{job: job, stop: make(chan struct{})})
	return nil
}

// Start begins executing all registered jobs in background goroutines.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.jobs {
		if !e.running {
			e.running = true
			go s.run(ctx, e)
		}
	}
}

func (s *Scheduler) run(ctx context.Context, e *jobEntry) {
	ticker := time.NewTicker(e.job.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = e.job.Task(ctx)
		case <-e.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop halts all running jobs.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.jobs {
		if e.running {
			close(e.stop)
			e.running = false
		}
	}
}

// JobNames returns the names of all registered jobs.
func (s *Scheduler) JobNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, len(s.jobs))
	for i, e := range s.jobs {
		names[i] = e.job.Name
	}
	return names
}
