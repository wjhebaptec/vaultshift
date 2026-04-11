package throttle

import (
	"context"
	"fmt"
)

// ProviderFunc is a generic function signature matching secret provider operations.
type ProviderFunc func(ctx context.Context, key string) (string, error)

// PutFunc is a generic function signature for put/write operations.
type PutFunc func(ctx context.Context, key, value string) error

// WrapGet wraps a ProviderFunc with throttle enforcement for the given key scope.
// The scope is prepended to the key to allow per-provider rate limiting.
func WrapGet(t *Throttler, scope string, fn ProviderFunc) ProviderFunc {
	return func(ctx context.Context, key string) (string, error) {
		tk := throttleKey(scope, key)
		if err := t.Allow(ctx, tk); err != nil {
			return "", fmt.Errorf("get throttled: %w", err)
		}
		return fn(ctx, key)
	}
}

// WrapPut wraps a PutFunc with throttle enforcement for the given key scope.
func WrapPut(t *Throttler, scope string, fn PutFunc) PutFunc {
	return func(ctx context.Context, key, value string) error {
		tk := throttleKey(scope, key)
		if err := t.Allow(ctx, tk); err != nil {
			return fmt.Errorf("put throttled: %w", err)
		}
		return fn(ctx, key, value)
	}
}

func throttleKey(scope, key string) string {
	if scope == "" {
		return key
	}
	return scope + ":" + key
}
