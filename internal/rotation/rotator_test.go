package rotation

import (
	"context"
	"testing"

	"github.com/yourusername/vaultshift/internal/config"
	"github.com/yourusername/vaultshift/internal/provider"
	"github.com/yourusername/vaultshift/internal/provider/mock"
)

func setupRotator(t *testing.T) (*Rotator, *provider.Registry) {
	t.Helper()

	registry := provider.NewRegistry()

	// Register source and target providers
	sourceProvider := mock.New("source")
	target1Provider := mock.New("target1")
	target2Provider := mock.New("target2")

	registry.Register("source", sourceProvider)
	registry.Register("target1", target1Provider)
	registry.Register("target2", target2Provider)

	cfg := &config.Config{
		Version: "1.0",
		Source: config.ProviderConfig{
			Name: "source",
			Type: "mock",
		},
		Targets: []config.ProviderConfig{
			{Name: "target1", Type: "mock"},
			{Name: "target2", Type: "mock"},
		},
		SecretKeys: []string{"api-key", "db-password"},
	}

	rotator := New(registry, cfg)
	return rotator, registry
}

func TestRotator_Rotate(t *testing.T) {
	rotator, registry := setupRotator(t)
	ctx := context.Background()

	// Setup source secret
	sourceProvider, _ := registry.Get("source")
	err := sourceProvider.PutSecret(ctx, "api-key", "secret-value-123")
	if err != nil {
		t.Fatalf("failed to put source secret: %v", err)
	}

	results, err := rotator.Rotate(ctx, "api-key")
	if err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if !result.Success {
			t.Errorf("rotation to %s failed: %v", result.TargetName, result.Error)
		}
	}

	// Verify secrets were synced
	target1Provider, _ := registry.Get("target1")
	value, err := target1Provider.GetSecret(ctx, "api-key")
	if err != nil {
		t.Errorf("failed to get secret from target1: %v", err)
	}
	if value != "secret-value-123" {
		t.Errorf("expected 'secret-value-123', got '%s'", value)
	}
}

func TestRotator_RotateAll(t *testing.T) {
	rotator, registry := setupRotator(t)
	ctx := context.Background()

	// Setup source secrets
	sourceProvider, _ := registry.Get("source")
	sourceProvider.PutSecret(ctx, "api-key", "key-123")
	sourceProvider.PutSecret(ctx, "db-password", "pass-456")

	results, err := rotator.RotateAll(ctx)
	if err != nil {
		t.Fatalf("RotateAll failed: %v", err)
	}

	// Should have 4 results (2 keys * 2 targets)
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
}
