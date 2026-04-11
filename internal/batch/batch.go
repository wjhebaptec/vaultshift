// Package batch provides utilities for processing secrets in configurable batches.
package batch

import (
	"context"
	"fmt"
	"sync"
)

// Item represents a single unit of work in a batch.
type Item struct {
	Key   string
	Value string
	Meta  map[string]string
}

// Result holds the outcome of processing a single Item.
type Result struct {
	Key   string
	Err   error
}

// ProcessFunc is the function applied to each Item during batch execution.
type ProcessFunc func(ctx context.Context, item Item) error

// Processor executes work items in bounded batches with optional concurrency.
type Processor struct {
	size    int
	workers int
}

// Option configures a Processor.
type Option func(*Processor)

// WithSize sets the maximum number of items per batch.
func WithSize(n int) Option {
	return func(p *Processor) {
		if n > 0 {
			p.size = n
		}
	}
}

// WithWorkers sets the number of concurrent workers per batch.
func WithWorkers(n int) Option {
	return func(p *Processor) {
		if n > 0 {
			p.workers = n
		}
	}
}

// New creates a Processor with the given options.
func New(opts ...Option) *Processor {
	p := &Processor{size: 10, workers: 1}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Run processes all items using fn, returning one Result per item.
func (p *Processor) Run(ctx context.Context, items []Item, fn ProcessFunc) []Result {
	results := make([]Result, len(items))
	for i := 0; i < len(items); i += p.size {
		end := i + p.size
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]
		p.processBatch(ctx, chunk, items, i, results, fn)
	}
	return results
}

func (p *Processor) processBatch(ctx context.Context, chunk []Item, _ []Item, offset int, results []Result, fn ProcessFunc) {
	sem := make(chan struct{}, p.workers)
	var wg sync.WaitGroup
	for j, item := range chunk {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, it Item) {
			defer wg.Done()
			defer func() { <-sem }()
			var err error
			select {
			case <-ctx.Done():
				err = fmt.Errorf("batch: context cancelled for key %q: %w", it.Key, ctx.Err())
			default:
				err = fn(ctx, it)
			}
			results[offset+idx] = Result{Key: it.Key, Err: err}
		}(j, item)
	}
	wg.Wait()
}
