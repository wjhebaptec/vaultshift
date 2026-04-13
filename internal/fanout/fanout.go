// Package fanout provides a mechanism for writing a secret to multiple
// destination providers simultaneously, collecting any errors that occur.
package fanout

import (
	"context"
	"fmt"
	"sync"

	"github.com/vaultshift/internal/provider"
)

// Result holds the outcome of a single fanout write operation.
type Result struct {
	Provider string
	Key      string
	Err      error
}

// Fanout writes secrets to multiple providers concurrently.
type Fanout struct {
	registry *provider.Registry
	targets  []string
}

// New creates a Fanout that will broadcast writes to the named targets.
func New(registry *provider.Registry, targets []string) (*Fanout, error) {
	if registry == nil {
		return nil, fmt.Errorf("fanout: registry must not be nil")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("fanout: at least one target provider is required")
	}
	return &Fanout{registry: registry, targets: targets}, nil
}

// Put writes key/value to all target providers concurrently and returns one
// Result per target.
func (f *Fanout) Put(ctx context.Context, key, value string) []Result {
	results := make([]Result, len(f.targets))
	var wg sync.WaitGroup

	for i, name := range f.targets {
		wg.Add(1)
		go func(idx int, providerName string) {
			defer wg.Done()
			p, err := f.registry.Get(providerName)
			if err != nil {
				results[idx] = Result{Provider: providerName, Key: key, Err: err}
				return
			}
			results[idx] = Result{
				Provider: providerName,
				Key:      key,
				Err:      p.Put(ctx, key, value),
			}
		}(i, name)
	}

	wg.Wait()
	return results
}

// HasFailures returns true if any Result contains a non-nil error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
