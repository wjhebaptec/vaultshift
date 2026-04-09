package rotation

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Strategy defines how secrets should be rotated
type Strategy interface {
	// Generate creates a new secret value
	Generate(ctx context.Context) (string, error)
}

// CopyStrategy copies the existing secret value from source
type CopyStrategy struct{}

// NewCopyStrategy creates a new CopyStrategy
func NewCopyStrategy() *CopyStrategy {
	return &CopyStrategy{}
}

// Generate returns empty string as CopyStrategy doesn't generate new values
func (s *CopyStrategy) Generate(ctx context.Context) (string, error) {
	return "", nil
}

// RandomStrategy generates a new random secret value
type RandomStrategy struct {
	length int
}

// NewRandomStrategy creates a new RandomStrategy with specified length
func NewRandomStrategy(length int) *RandomStrategy {
	if length <= 0 {
		length = 32 // default length
	}
	return &RandomStrategy{length: length}
}

// Generate creates a new random base64-encoded secret
func (s *RandomStrategy) Generate(ctx context.Context) (string, error) {
	bytes := make([]byte, s.length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CustomStrategy allows using a custom function to generate secrets
type CustomStrategy struct {
	generateFn func(context.Context) (string, error)
}

// NewCustomStrategy creates a new CustomStrategy
func NewCustomStrategy(fn func(context.Context) (string, error)) *CustomStrategy {
	return &CustomStrategy{generateFn: fn}
}

// Generate calls the custom generation function
func (s *CustomStrategy) Generate(ctx context.Context) (string, error) {
	if s.generateFn == nil {
		return "", fmt.Errorf("generate function not set")
	}
	return s.generateFn(ctx)
}

// RotateWithStrategy rotates a secret using the specified strategy
func (r *Rotator) RotateWithStrategy(ctx context.Context, secretKey string, strategy Strategy) ([]RotateResult, error) {
	newValue, err := strategy.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new secret: %w", err)
	}

	// If strategy generates empty value, use copy strategy (fetch from source)
	if newValue == "" {
		return r.Rotate(ctx, secretKey)
	}

	// Put new value to source first
	sourceProvider, err := r.registry.Get(r.config.Source.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get source provider: %w", err)
	}

	if err := sourceProvider.PutSecret(ctx, secretKey, newValue); err != nil {
		return nil, fmt.Errorf("failed to update source secret: %w", err)
	}

	// Then rotate to targets
	return r.Rotate(ctx, secretKey)
}
