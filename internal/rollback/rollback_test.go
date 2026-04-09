package rollback_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/rollback"
	"github.com/vaultshift/internal/version"
)

func setupRollbacker(t *testing.T) (*rollback.Rollbacker, *mock.Provider, *version.History) {
	t.Helper()
	mp := mock.New()
	reg := provider.NewRegistry()
	reg.Register("test", mp)
	h := version.NewHistory(10)
	return rollback.New(reg, h), mp, h
}

func TestRollback_RestoresPreviousValue(t *testing.T) {
	rb, mp, h := setupRollbacker(t)
	ctx := context.Background()

	h.Push("db/password", "old-secret")
	h.Push("db/password", "new-secret")
	_ = mp.Put(ctx, "db/password", "new-secret")

	rec, err := rb.Rollback(ctx, "test", "db/password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.PrevValue != "old-secret" {
		t.Errorf("expected prev value %q, got %q", "old-secret", rec.PrevValue)
	}

	val, _ := mp.Get(ctx, "db/password")
	if val != "old-secret" {
		t.Errorf("provider should hold %q, got %q", "old-secret", val)
	}
}

func TestRollback_UnknownProvider(t *testing.T) {
	rb, _, h := setupRollbacker(t)
	h.Push("key", "v1")
	h.Push("key", "v2")

	_, err := rb.Rollback(context.Background(), "nonexistent", "key")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRollback_NoPreviousVersion(t *testing.T) {
	rb, _, h := setupRollbacker(t)
	h.Push("key", "only-version")

	_, err := rb.Rollback(context.Background(), "test", "key")
	if err == nil {
		t.Fatal("expected error when no previous version exists")
	}
}

func TestRollbackAll_RestoresMultipleKeys(t *testing.T) {
	rb, mp, h := setupRollbacker(t)
	ctx := context.Background()

	for _, k := range []string{"a", "b"} {
		h.Push(k, "old-"+k)
		h.Push(k, "new-"+k)
		_ = mp.Put(ctx, k, "new-"+k)
	}

	records, err := rb.RollbackAll(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}
