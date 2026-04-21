// Package gradient provides a weighted provider selector that routes
// secret operations across multiple backends according to configured
// traffic weights, enabling canary rollouts and gradual migrations.
package gradient

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/vaultshift/internal/provider"
)

// ErrNoProviders is returned when the gradient has no registered providers.
var ErrNoProviders = errors.New("gradient: no providers registered")

// ErrWeightMismatch is returned when provider and weight slices differ in length.
var ErrWeightMismatch = errors.New("gradient: provider and weight counts must match")

// ErrNegativeWeight is returned when any weight is negative.
var ErrNegativeWeight = errors.New("gradient: weights must be non-negative")

// entry holds a provider alongside its cumulative weight boundary.
type entry struct {
	name     string
	p        provider.Provider
	upper    float64
}

// Gradient routes operations to one of several providers based on weights.
type Gradient struct {
	mu      sync.RWMutex
	entries []entry
	total   float64
	rand    func() float64
}

// New constructs a Gradient from named providers and corresponding weights.
// Weights must be non-negative and at least one must be non-zero.
func New(reg *provider.Registry, names []string, weights []float64) (*Gradient, error) {
	if len(names) == 0 {
		return nil, ErrNoProviders
	}
	if len(names) != len(weights) {
		return nil, ErrWeightMismatch
	}

	var total float64
	for _, w := range weights {
		if w < 0 {
			return nil, ErrNegativeWeight
		}
		total += w
	}
	if total == 0 {
		return nil, errors.New("gradient: total weight must be greater than zero")
	}

	entries := make([]entry, len(names))
	var cumulative float64
	for i, name := range names {
		p, ok := reg.Get(name)
		if !ok {
			return nil, fmt.Errorf("gradient: provider %q not found", name)
		}
		cumulative += weights[i]
		entries[i] = entry{name: name, p: p, upper: cumulative / total}
	}

	return &Gradient{
		entries: entries,
		total:   total,
		rand:    rand.Float64,
	}, nil
}

// pick selects a provider according to the weight distribution.
func (g *Gradient) pick() (provider.Provider, string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r := g.rand()
	for _, e := range g.entries {
		if r <= e.upper {
			return e.p, e.name
		}
	}
	last := g.entries[len(g.entries)-1]
	return last.p, last.name
}

// Get retrieves a secret from the selected provider.
func (g *Gradient) Get(key string) (string, error) {
	p, _ := g.pick()
	return p.Get(key)
}

// Put writes a secret to the selected provider.
func (g *Gradient) Put(key, value string) error {
	p, _ := g.pick()
	return p.Put(key, value)
}

// Selected returns the name of the provider that would be chosen for the next
// operation (useful for inspection / testing with a fixed rand source).
func (g *Gradient) Selected() string {
	_, name := g.pick()
	return name
}
