package sync

import (
	"context"
	"testing"

	"vaultshift/internal/provider"
	"vaultshift/internal/provider/mock"
)

func setupSyncer(t *testing.T) (*Syncer, *provider.Registry) {
	t.Helper()
	registry := provider.NewRegistry()

	sourceMock := mock.New()
	target1Mock := mock.New()
	target2Mock := mock.New()

	registry.Register("source", sourceMock)
	registry.Register("target1", target1Mock)
	registry.Register("target2", target2Mock)

	return New(registry), registry
}

func TestSyncer_SyncSecret(t *testing.T) {
	syncer, registry := setupSyncer(t)
	ctx := context.Background()

	source, _ := registry.GetProvider("source")
	_ = source.PutSecret(ctx, "api-key", "secret-value")

	err := syncer.SyncSecret(ctx, "api-key", "source", []string{"target1", "target2"})
	if err != nil {
		t.Fatalf("SyncSecret failed: %v", err)
	}

	target1, _ := registry.GetProvider("target1")
	value1, err := target1.GetSecret(ctx, "api-key")
	if err != nil || value1 != "secret-value" {
		t.Errorf("target1 did not receive secret correctly: got %v, err %v", value1, err)
	}

	target2, _ := registry.GetProvider("target2")
	value2, err := target2.GetSecret(ctx, "api-key")
	if err != nil || value2 != "secret-value" {
		t.Errorf("target2 did not receive secret correctly: got %v, err %v", value2, err)
	}
}

func TestSyncer_SyncAll(t *testing.T) {
	syncer, registry := setupSyncer(t)
	ctx := context.Background()

	source, _ := registry.GetProvider("source")
	_ = source.PutSecret(ctx, "key1", "value1")
	_ = source.PutSecret(ctx, "key2", "value2")

	err := syncer.SyncAll(ctx, "source", []string{"target1"})
	if err != nil {
		t.Fatalf("SyncAll failed: %v", err)
	}

	target1, _ := registry.GetProvider("target1")
	value1, _ := target1.GetSecret(ctx, "key1")
	value2, _ := target1.GetSecret(ctx, "key2")

	if value1 != "value1" || value2 != "value2" {
		t.Errorf("SyncAll did not sync all secrets correctly")
	}
}

func TestSyncer_SyncWithFilter(t *testing.T) {
	syncer, registry := setupSyncer(t)
	ctx := context.Background()

	source, _ := registry.GetProvider("source")
	_ = source.PutSecret(ctx, "prod/key1", "value1")
	_ = source.PutSecret(ctx, "dev/key2", "value2")
	_ = source.PutSecret(ctx, "prod/key3", "value3")

	filter := NewPrefixFilter("prod/")
	err := syncer.SyncWithFilter(ctx, "source", []string{"target1"}, filter)
	if err != nil {
		t.Fatalf("SyncWithFilter failed: %v", err)
	}

	target1, _ := registry.GetProvider("target1")
	_, err1 := target1.GetSecret(ctx, "prod/key1")
	_, err2 := target1.GetSecret(ctx, "dev/key2")
	_, err3 := target1.GetSecret(ctx, "prod/key3")

	if err1 != nil || err3 != nil {
		t.Errorf("prod secrets should be synced")
	}
	if err2 == nil {
		t.Errorf("dev secret should not be synced")
	}
}
