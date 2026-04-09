package mock

import (
	"context"
	"testing"

	"vaultshift/internal/provider"
)

func TestMockProvider_PutAndGet(t *testing.T) {
	p := New(provider.TypeAWS)
	ctx := context.Background()

	secret := &provider.Secret{
		Key:     "api-key",
		Value:   "secret-value",
		Version: "v1",
	}

	if err := p.PutSecret(ctx, secret); err != nil {
		t.Fatalf("PutSecret failed: %v", err)
	}

	retrieved, err := p.GetSecret(ctx, "api-key")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}

	if retrieved.Value != "secret-value" {
		t.Errorf("expected value %s, got %s", "secret-value", retrieved.Value)
	}
}

func TestMockProvider_GetNonExistent(t *testing.T) {
	p := New(provider.TypeGCP)
	ctx := context.Background()

	_, err := p.GetSecret(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent secret")
	}
}

func TestMockProvider_Delete(t *testing.T) {
	p := New(provider.TypeVault)
	ctx := context.Background()

	secret := &provider.Secret{Key: "temp", Value: "data"}
	p.PutSecret(ctx, secret)

	if err := p.DeleteSecret(ctx, "temp"); err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}

	_, err := p.GetSecret(ctx, "temp")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestMockProvider_ListSecrets(t *testing.T) {
	p := New(provider.TypeAWS)
	ctx := context.Background()

	p.PutSecret(ctx, &provider.Secret{Key: "key1", Value: "val1"})
	p.PutSecret(ctx, &provider.Secret{Key: "key2", Value: "val2"})

	keys, err := p.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestMockProvider_ClosedOperations(t *testing.T) {
	p := New(provider.TypeAWS)
	ctx := context.Background()

	p.Close()

	if err := p.PutSecret(ctx, &provider.Secret{Key: "key", Value: "val"}); err == nil {
		t.Error("expected error on closed provider")
	}
}
