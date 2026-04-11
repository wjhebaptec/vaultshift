package circuitbreaker

import (
	"context"
	"fmt"
)

// Provider is the minimal interface required to wrap with a circuit breaker.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
}

// WrappedProvider wraps a Provider with circuit breaker logic.
type WrappedProvider struct {
	inner   Provider
	cb      *CircuitBreaker
	provider string
}

// Wrap returns a WrappedProvider that guards the given provider with the circuit breaker.
func Wrap(provider string, inner Provider, cb *CircuitBreaker) *WrappedProvider {
	return &WrappedProvider{inner: inner, cb: cb, provider: provider}
}

// Get retrieves a secret value, respecting the circuit breaker state.
func (w *WrappedProvider) Get(ctx context.Context, key string) (string, error) {
	if err := w.cb.Allow(); err != nil {
		return "", fmt.Errorf("provider %s: %w", w.provider, err)
	}
	val, err := w.inner.Get(ctx, key)
	if err != nil {
		w.cb.RecordFailure()
		return "", err
	}
	w.cb.RecordSuccess()
	return val, nil
}

// Put stores a secret value, respecting the circuit breaker state.
func (w *WrappedProvider) Put(ctx context.Context, key, value string) error {
	if err := w.cb.Allow(); err != nil {
		return fmt.Errorf("provider %s: %w", w.provider, err)
	}
	if err := w.inner.Put(ctx, key, value); err != nil {
		w.cb.RecordFailure()
		return err
	}
	w.cb.RecordSuccess()
	return nil
}

// Breaker returns the underlying CircuitBreaker for inspection.
func (w *WrappedProvider) Breaker() *CircuitBreaker {
	return w.cb
}
