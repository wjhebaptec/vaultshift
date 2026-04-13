// Package stagger provides time-staggered execution of operations across
// multiple targets to avoid thundering-herd problems when rotating or
// syncing secrets at scale.
package stagger

import (
	"context"
	"fmt"
	"time"
)

// Task is a named unit of work that can be staggered.
type Task struct {
	Name string
	Fn   func(ctx context.Context) error
}

// Result holds the outcome of a single staggered task.
type Result struct {
	Name    string
	Err     error
	Elapsed time.Duration
}

// Runner executes tasks with a configurable delay between each.
type Runner struct {
	delay   time.Duration
	tasks   []Task
	clock   func() time.Time
}

// Option configures a Runner.
type Option func(*Runner)

// WithDelay sets the inter-task delay.
func WithDelay(d time.Duration) Option {
	return func(r *Runner) { r.delay = d }
}

// New creates a Runner with the provided options.
func New(opts ...Option) *Runner {
	r := &Runner{
		delay: 500 * time.Millisecond,
		clock: time.Now,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Add appends a task to the runner.
func (r *Runner) Add(name string, fn func(ctx context.Context) error) error {
	if name == "" {
		return fmt.Errorf("stagger: task name must not be empty")
	}
	if fn == nil {
		return fmt.Errorf("stagger: task fn must not be nil")
	}
	r.tasks = append(r.tasks, Task{Name: name, Fn: fn})
	return nil
}

// Run executes all tasks sequentially, sleeping for the configured delay
// between each. Execution stops early if ctx is cancelled.
func (r *Runner) Run(ctx context.Context) []Result {
	results := make([]Result, 0, len(r.tasks))
	for i, t := range r.tasks {
		if i > 0 {
			select {
			case <-ctx.Done():
				results = append(results, Result{Name: t.Name, Err: ctx.Err()})
				continue
			case <-time.After(r.delay):
			}
		}
		start := r.clock()
		err := t.Fn(ctx)
		results = append(results, Result{
			Name:    t.Name,
			Err:     err,
			Elapsed: r.clock().Sub(start),
		})
	}
	return results
}

// HasFailures returns true if any result contains a non-nil error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
