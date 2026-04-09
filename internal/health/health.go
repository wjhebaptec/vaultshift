// Package health provides provider connectivity checks for vaultshift.
package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents the health state of a provider.
type Status string

const (
	StatusOK      Status = "ok"
	StatusDegraded Status = "degraded"
	StatusUnknown  Status = "unknown"
)

// Result holds the health check outcome for a single provider.
type Result struct {
	Provider  string
	Status    Status
	Latency   time.Duration
	Err       error
	CheckedAt time.Time
}

// Checker defines the interface a provider must satisfy for health checks.
type Checker interface {
	Name() string
	Ping(ctx context.Context) error
}

// Monitor runs health checks against registered providers.
type Monitor struct {
	mu       sync.Mutex
	checkers []Checker
}

// New creates a new Monitor.
func New() *Monitor {
	return &Monitor{}
}

// Register adds a Checker to the monitor.
func (m *Monitor) Register(c Checker) error {
	if c == nil {
		return fmt.Errorf("health: checker must not be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers = append(m.checkers, c)
	return nil
}

// CheckAll runs Ping on every registered checker and returns all results.
func (m *Monitor) CheckAll(ctx context.Context) []Result {
	m.mu.Lock()
	checkers := make([]Checker, len(m.checkers))
	copy(checkers, m.checkers)
	m.mu.Unlock()

	results := make([]Result, 0, len(checkers))
	for _, c := range checkers {
		results = append(results, m.check(ctx, c))
	}
	return results
}

// Check runs a single Ping by provider name.
func (m *Monitor) Check(ctx context.Context, name string) (Result, error) {
	m.mu.Lock()
	var target Checker
	for _, c := range m.checkers {
		if c.Name() == name {
			target = c
			break
		}
	}
	m.mu.Unlock()

	if target == nil {
		return Result{}, fmt.Errorf("health: no checker registered for %q", name)
	}
	return m.check(ctx, target), nil
}

func (m *Monitor) check(ctx context.Context, c Checker) Result {
	start := time.Now()
	err := c.Ping(ctx)
	latency := time.Since(start)

	status := StatusOK
	if err != nil {
		status = StatusDegraded
	}
	return Result{
		Provider:  c.Name(),
		Status:    status,
		Latency:   latency,
		Err:       err,
		CheckedAt: time.Now(),
	}
}
