package gradient

import (
	"errors"
	"fmt"
	"sync"
)

// Shift describes a single weight reassignment.
type Shift struct {
	Name   string
	Weight float64
}

// Shifter manages live weight updates to a Gradient without reconstruction.
type Shifter struct {
	mu       sync.Mutex
	gradient *Gradient
}

// NewShifter wraps an existing Gradient for dynamic weight management.
func NewShifter(g *Gradient) (*Shifter, error) {
	if g == nil {
		return nil, errors.New("gradient: shifter requires a non-nil gradient")
	}
	return &Shifter{gradient: g}, nil
}

// Apply updates the weight distribution of the underlying Gradient.
// The shifts slice must contain exactly one entry per registered provider
// and all weights must be non-negative with a positive total.
func (s *Shifter) Apply(shifts []Shift) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	g := s.gradient
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(shifts) != len(g.entries) {
		return fmt.Errorf("gradient: shift count %d does not match provider count %d",
			len(shifts), len(g.entries))
	}

	// Build name→index map for validation.
	index := make(map[string]int, len(g.entries))
	for i, e := range g.entries {
		index[e.name] = i
	}

	weights := make([]float64, len(g.entries))
	var total float64
	for _, sh := range shifts {
		i, ok := index[sh.Name]
		if !ok {
			return fmt.Errorf("gradient: unknown provider %q in shift", sh.Name)
		}
		if sh.Weight < 0 {
			return ErrNegativeWeight
		}
		weights[i] = sh.Weight
		total += sh.Weight
	}
	if total == 0 {
		return errors.New("gradient: total weight after shift must be greater than zero")
	}

	var cumulative float64
	for i := range g.entries {
		cumulative += weights[i]
		g.entries[i].upper = cumulative / total
	}
	g.total = total
	return nil
}

// Weights returns a snapshot of the current name→weight mapping.
func (s *Shifter) Weights() map[string]float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	g := s.gradient
	g.mu.RLock()
	defer g.mu.RUnlock()

	out := make(map[string]float64, len(g.entries))
	var prev float64
	for _, e := range g.entries {
		out[e.name] = (e.upper - prev) * g.total
		prev = e.upper
	}
	return out
}
