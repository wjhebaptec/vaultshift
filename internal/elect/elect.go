// Package elect provides a simple leader-election mechanism for
// coordinating which instance of vaultshift is responsible for
// performing a rotation or sync operation at any given time.
package elect

import (
	"errors"
	"sync"
	"time"
)

// ErrAlreadyLeader is returned when the caller is already the leader.
var ErrAlreadyLeader = errors.New("elect: already leader")

// ErrNotLeader is returned when the caller is not the current leader.
var ErrNotLeader = errors.New("elect: not the leader")

// ErrLeaderExists is returned when another candidate holds the lease.
var ErrLeaderExists = errors.New("elect: another leader holds the lease")

// Elector manages a single-leader lease.
type Elector struct {
	mu       sync.Mutex
	leader   string
	leaseEnd time.Time
	leaseTTL time.Duration
	clock    func() time.Time
}

// Option configures an Elector.
type Option func(*Elector)

// WithTTL sets the lease duration.
func WithTTL(d time.Duration) Option {
	return func(e *Elector) { e.leaseTTL = d }
}

// WithClock overrides the time source (useful in tests).
func WithClock(fn func() time.Time) Option {
	return func(e *Elector) { e.clock = fn }
}

// New creates an Elector with the given options.
func New(opts ...Option) (*Elector, error) {
	e := &Elector{
		leaseTTL: 30 * time.Second,
		clock:    time.Now,
	}
	for _, o := range opts {
		o(e)
	}
	if e.leaseTTL <= 0 {
		return nil, errors.New("elect: leaseTTL must be positive")
	}
	return e, nil
}

// Campaign attempts to acquire leadership for the given candidate.
// Returns ErrLeaderExists if another candidate currently holds the lease.
func (e *Elector) Campaign(candidate string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.clock()
	if e.leader != "" && now.Before(e.leaseEnd) {
		if e.leader == candidate {
			return ErrAlreadyLeader
		}
		return ErrLeaderExists
	}
	e.leader = candidate
	e.leaseEnd = now.Add(e.leaseTTL)
	return nil
}

// Renew extends the lease for the current leader.
func (e *Elector) Renew(candidate string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leader != candidate {
		return ErrNotLeader
	}
	e.leaseEnd = e.clock().Add(e.leaseTTL)
	return nil
}

// Resign releases the lease held by the given candidate.
func (e *Elector) Resign(candidate string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leader != candidate {
		return ErrNotLeader
	}
	e.leader = ""
	e.leaseEnd = time.Time{}
	return nil
}

// Leader returns the current leader and whether the lease is active.
func (e *Elector) Leader() (string, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leader == "" || e.clock().After(e.leaseEnd) {
		return "", false
	}
	return e.leader, true
}
