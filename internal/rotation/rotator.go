package rotation

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/provider"
)

// Rotator handles secret rotation across providers
type Rotator struct {
	registry *provider.Registry
	config   *config.Config
}

// New creates a new Rotator instance
func New(registry *provider.Registry, cfg *config.Config) *Rotator {
	return &Rotator{
		registry: registry,
		config:   cfg,
	}
}

// RotateResult contains the result of a rotation operation
type RotateResult struct {
	SecretKey  string
	SourceName string
	TargetName string
	Success    bool
	Error      error
	RotatedAt  time.Time
}

// Rotate performs secret rotation from source to targets
func (r *Rotator) Rotate(ctx context.Context, secretKey string) ([]RotateResult, error) {
	sourceProvider, err := r.registry.Get(r.config.Source.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get source provider: %w", err)
	}

	// Get secret from source
	secretValue, err := sourceProvider.GetSecret(ctx, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from source: %w", err)
	}

	results := make([]RotateResult, 0, len(r.config.Targets))

	// Sync to all targets
	for _, target := range r.config.Targets {
		result := RotateResult{
			SecretKey:  secretKey,
			SourceName: r.config.Source.Name,
			TargetName: target.Name,
			RotatedAt:  time.Now(),
		}

		targetProvider, err := r.registry.Get(target.Name)
		if err != nil {
			result.Error = fmt.Errorf("failed to get target provider: %w", err)
			results = append(results, result)
			continue
		}

		if err := targetProvider.PutSecret(ctx, secretKey, secretValue); err != nil {
			result.Error = fmt.Errorf("failed to put secret: %w", err)
			results = append(results, result)
			continue
		}

		result.Success = true
		results = append(results, result)
	}

	return results, nil
}

// RotateAll rotates all configured secret keys
func (r *Rotator) RotateAll(ctx context.Context) ([]RotateResult, error) {
	if len(r.config.SecretKeys) == 0 {
		return nil, fmt.Errorf("no secret keys configured")
	}

	allResults := make([]RotateResult, 0)
	for _, key := range r.config.SecretKeys {
		results, err := r.Rotate(ctx, key)
		if err != nil {
			return allResults, fmt.Errorf("failed to rotate key %s: %w", key, err)
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// HasFailures returns true if any result in the slice represents a failed rotation.
// This is useful for callers that need to check overall success without iterating
// results manually.
func HasFailures(results []RotateResult) bool {
	for _, r := range results {
		if !r.Success {
			return true
		}
	}
	return false
}
