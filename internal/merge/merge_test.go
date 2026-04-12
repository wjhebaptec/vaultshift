package merge_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/merge"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupMerger(t *testing.T, strategy merge.Strategy) (*merge.Merger, *provider.Registry) {
	t.Helper()
	reg := provider.NewRegistry()

	a := mock.New()
	_ = a.PutSecret(context.Background(), "shared", "from-a")
	_ = a.PutSecret(context.Background(), "only-a", "value-a")
	reg.Register("a", a)

	b := mock.New()
	_ = b.PutSecret(context.Background(), "shared", "from-b")
	_ = b.PutSecret(context.Background(), "only-b", "value-b")
	reg.Register("b", b)

	return merge.New(reg, merge.WithStrategy(strategy)), reg
}

func TestMerge_StrategyFirst_KeepsFirstValue(t *testing.T) {
	m, _ := setupMerger(t, merge.StrategyFirst)
	result, err := m.Merge(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["shared"] != "from-a" {
		t.Errorf("expected from-a, got %q", result["shared"])
	}
	if result["only-b"] != "value-b" {
		t.Errorf("expected value-b, got %q", result["only-b"])
	}
}

func TestMerge_StrategyLast_OverwritesWithLast(t *testing.T) {
	m, _ := setupMerger(t, merge.StrategyLast)
	result, err := m.Merge(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["shared"] != "from-b" {
		t.Errorf("expected from-b, got %q", result["shared"])
	}
}

func TestMerge_StrategyError_ReturnsErrorOnConflict(t *testing.T) {
	m, _ := setupMerger(t, merge.StrategyError)
	_, err := m.Merge(context.Background(), "a", "b")
	if err == nil {
		t.Fatal("expected error for conflicting key, got nil")
	}
}

func TestMerge_NoConflict_CombinesAll(t *testing.T) {
	reg := provider.NewRegistry()
	a := mock.New()
	_ = a.PutSecret(context.Background(), "key-a", "val-a")
	reg.Register("a", a)
	b := mock.New()
	_ = b.PutSecret(context.Background(), "key-b", "val-b")
	reg.Register("b", b)

	m := merge.New(reg)
	result, err := m.Merge(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
}

func TestMerge_UnknownProvider_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	m := merge.New(reg)
	_, err := m.Merge(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}
