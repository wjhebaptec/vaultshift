package rotation

import (
	"context"
	"errors"
	"testing"
)

func TestCopyStrategy_Generate(t *testing.T) {
	strategy := NewCopyStrategy()
	ctx := context.Background()

	value, err := strategy.Generate(ctx)
	if err != nil {
		t.Errorf("CopyStrategy.Generate() error = %v", err)
	}
	if value != "" {
		t.Errorf("expected empty string, got %s", value)
	}
}

func TestRandomStrategy_Generate(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"default length", 0},
		{"custom length 16", 16},
		{"custom length 64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewRandomStrategy(tt.length)
			ctx := context.Background()

			value, err := strategy.Generate(ctx)
			if err != nil {
				t.Errorf("RandomStrategy.Generate() error = %v", err)
			}
			if value == "" {
				t.Error("expected non-empty value")
			}

			// Generate again to ensure uniqueness
			value2, err := strategy.Generate(ctx)
			if err != nil {
				t.Errorf("RandomStrategy.Generate() error = %v", err)
			}
			if value == value2 {
				t.Error("expected different random values")
			}
		})
	}
}

func TestCustomStrategy_Generate(t *testing.T) {
	t.Run("with valid function", func(t *testing.T) {
		expected := "custom-secret-123"
		strategy := NewCustomStrategy(func(ctx context.Context) (string, error) {
			return expected, nil
		})

		value, err := strategy.Generate(context.Background())
		if err != nil {
			t.Errorf("CustomStrategy.Generate() error = %v", err)
		}
		if value != expected {
			t.Errorf("expected %s, got %s", expected, value)
		}
	})

	t.Run("with error function", func(t *testing.T) {
		expectedErr := errors.New("generation failed")
		strategy := NewCustomStrategy(func(ctx context.Context) (string, error) {
			return "", expectedErr
		})

		_, err := strategy.Generate(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("with nil function", func(t *testing.T) {
		strategy := NewCustomStrategy(nil)

		_, err := strategy.Generate(context.Background())
		if err == nil {
			t.Error("expected error for nil function, got nil")
		}
	})
}

func TestRotator_RotateWithStrategy(t *testing.T) {
	rotator, registry := setupRotator(t)
	ctx := context.Background()

	// Setup initial source secret
	sourceProvider, _ := registry.Get("source")
	sourceProvider.PutSecret(ctx, "api-key", "old-value")

	strategy := NewRandomStrategy(32)
	results, err := rotator.RotateWithStrategy(ctx, "api-key", strategy)
	if err != nil {
		t.Fatalf("RotateWithStrategy failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify new value is different from old value
	newValue, _ := sourceProvider.GetSecret(ctx, "api-key")
	if newValue == "old-value" {
		t.Error("secret value was not rotated")
	}
}
