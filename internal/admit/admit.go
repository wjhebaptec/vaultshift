// Package admit provides a simple admission control layer that gates
// secret operations based on pluggable policy functions.
package admit

import (
	"context"
	"errors"
	"fmt"
)

// ErrDenied is returned when an admission check rejects an operation.
var ErrDenied = errors.New("admission denied")

// Op represents the type of secret operation being admitted.
type Op string

const (
	OpGet    Op = "get"
	OpPut    Op = "put"
	OpDelete Op = "delete"
	OpList   Op = "list"
)

// Request carries the context of an incoming operation.
type Request struct {
	Provider string
	Key      string
	Op       Op
}

// PolicyFunc is a function that evaluates an admission request.
// It returns nil to allow the request, or an error to deny it.
type PolicyFunc func(ctx context.Context, req Request) error

// Admitter evaluates requests against a set of registered policies.
type Admitter struct {
	policies []namedPolicy
}

type namedPolicy struct {
	name string
	fn   PolicyFunc
}

// New creates a new Admitter with no policies registered.
func New() *Admitter {
	return &Admitter{}
}

// Register adds a named policy to the admitter.
// Policies are evaluated in registration order.
func (a *Admitter) Register(name string, fn PolicyFunc) error {
	if name == "" {
		return errors.New("admit: policy name must not be empty")
	}
	if fn == nil {
		return errors.New("admit: policy function must not be nil")
	}
	for _, p := range a.policies {
		if p.name == name {
			return fmt.Errorf("admit: policy %q already registered", name)
		}
	}
	a.policies = append(a.policies, namedPolicy{name: name, fn: fn})
	return nil
}

// Admit evaluates all registered policies against the request.
// The first policy that returns an error causes Admit to return
// a wrapped ErrDenied.
func (a *Admitter) Admit(ctx context.Context, req Request) error {
	for _, p := range a.policies {
		if err := p.fn(ctx, req); err != nil {
			return fmt.Errorf("%w: policy %q rejected %s on %s/%s: %v",
				ErrDenied, p.name, req.Op, req.Provider, req.Key, err)
		}
	}
	return nil
}

// Len returns the number of registered policies.
func (a *Admitter) Len() int { return len(a.policies) }

// DenyAll is a convenience PolicyFunc that denies every request.
func DenyAll(_ context.Context, _ Request) error {
	return errors.New("all operations denied")
}

// AllowReadOnly is a convenience PolicyFunc that denies write operations.
func AllowReadOnly(_ context.Context, req Request) error {
	if req.Op == OpPut || req.Op == OpDelete {
		return fmt.Errorf("write operation %q is not permitted in read-only mode", req.Op)
	}
	return nil
}
