// Package overwrite provides a conditional write layer that only updates a
// secret when its value has actually changed, avoiding unnecessary writes to
// the underlying provider.
package overwrite

import (
	"context"
	"errors"
	"fmt"
)

// Provider is the minimal interface required by the overwrite layer.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]string, error)
}

// ErrNilProvider is returned when a nil provider is supplied.
var ErrNilProvider = errors.New("overwrite: provider must not be nil")

// Result describes the outcome of a conditional Put.
type Result struct {
	Key     string
	Written bool // false when the value was already current
}

// Guard wraps a Provider and skips Put calls when the stored value already
// matches the incoming value.
type Guard struct {
	inner Provider
}

// New returns a Guard wrapping inner. It returns ErrNilProvider when inner is nil.
func New(inner Provider) (*Guard, error) {
	if inner == nil {
		return nil, ErrNilProvider
	}
	return &Guard{inner: inner}, nil
}

// Put writes value to key only when the current stored value differs.
// It returns a Result indicating whether the write was performed.
func (g *Guard) Put(ctx context.Context, key, value string) (Result, error) {
	current, err := g.inner.Get(ctx, key)
	if err == nil && current == value {
		return Result{Key: key, Written: false}, nil
	}
	if err := g.inner.Put(ctx, key, value); err != nil {
		return Result{}, fmt.Errorf("overwrite: put %q: %w", key, err)
	}
	return Result{Key: key, Written: true}, nil
}

// Get delegates directly to the inner provider.
func (g *Guard) Get(ctx context.Context, key string) (string, error) {
	return g.inner.Get(ctx, key)
}

// Delete delegates directly to the inner provider.
func (g *Guard) Delete(ctx context.Context, key string) error {
	return g.inner.Delete(ctx, key)
}

// List delegates directly to the inner provider.
func (g *Guard) List(ctx context.Context) ([]string, error) {
	return g.inner.List(ctx)
}

// PutAll performs conditional writes for every entry in values.
// It returns all results and any accumulated errors.
func (g *Guard) PutAll(ctx context.Context, values map[string]string) ([]Result, error) {
	results := make([]Result, 0, len(values))
	var errs []error
	for k, v := range values {
		r, err := g.Put(ctx, k, v)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, r)
	}
	if len(errs) > 0 {
		return results, fmt.Errorf("overwrite: %d error(s): %v", len(errs), errs)
	}
	return results, nil
}
