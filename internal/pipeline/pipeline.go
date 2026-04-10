// Package pipeline provides a composable, sequential execution engine
// for chaining secret operations (rotate, sync, validate, audit).
package pipeline

import (
	"context"
	"fmt"
	"time"
)

// StepFunc is a single unit of work in a pipeline.
type StepFunc func(ctx context.Context, payload *Payload) error

// Payload carries state between pipeline steps.
type Payload struct {
	Key      string
	Provider string
	Value    string
	Meta     map[string]string
}

// Step wraps a StepFunc with a name for logging and metrics.
type Step struct {
	Name string
	Run  StepFunc
}

// Result holds the outcome of a single step execution.
type Result struct {
	Step     string
	Duration time.Duration
	Err      error
}

// Pipeline executes a sequence of Steps against a Payload.
type Pipeline struct {
	steps []Step
}

// New returns an empty Pipeline.
func New() *Pipeline {
	return &Pipeline{}
}

// Add appends a Step to the pipeline.
func (p *Pipeline) Add(s Step) *Pipeline {
	if s.Name == "" {
		s.Name = fmt.Sprintf("step-%d", len(p.steps)+1)
	}
	p.steps = append(p.steps, s)
	return p
}

// Execute runs all steps in order. It stops on the first error.
// It returns a slice of Results for observability.
func (p *Pipeline) Execute(ctx context.Context, payload *Payload) ([]Result, error) {
	results := make([]Result, 0, len(p.steps))
	for _, s := range p.steps {
		start := time.Now()
		err := s.Run(ctx, payload)
		results = append(results, Result{
			Step:     s.Name,
			Duration: time.Since(start),
			Err:      err,
		})
		if err != nil {
			return results, fmt.Errorf("pipeline step %q failed: %w", s.Name, err)
		}
		if err := ctx.Err(); err != nil {
			return results, fmt.Errorf("pipeline cancelled after step %q: %w", s.Name, err)
		}
	}
	return results, nil
}

// Len returns the number of registered steps.
func (p *Pipeline) Len() int { return len(p.steps) }
