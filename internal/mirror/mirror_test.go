package mirror_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/mirror"
	"github.com/vaultshift/internal/provider/mock"
)

func setup(t *testing.T) (context.Context, *mock.Provider, *mock.Provider, *mirror.Mirror) {
	t.Helper()
	ctx := context.Background()
	primary := mock.New()
	secondary := mock.New()
	m, err := mirror.New(primary, secondary)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return ctx, primary, secondary, m
}

func TestNew_NilPrimary_ReturnsError(t *testing.T) {
	_, err := mirror.New(nil, mock.New())
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNew_NilSecondary_ReturnsError(t *testing.T) {
	_, err := mirror.New(mock.New(), nil)
	if err == nil {
		t.Fatal("expected error for nil secondary")
	}
}

func TestGet_MirrorsToSecondary(t *testing.T) {
	ctx, primary, secondary, m := setup(t)
	_ = primary.Put(ctx, "db/pass", "secret123")

	val, err := m.Get(ctx, "db/pass")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "secret123" {
		t.Fatalf("want %q, got %q", "secret123", val)
	}

	mirrored, err := secondary.Get(ctx, "db/pass")
	if err != nil {
		t.Fatalf("secondary Get: %v", err)
	}
	if mirrored != "secret123" {
		t.Fatalf("secondary want %q, got %q", "secret123", mirrored)
	}
}

func TestPut_NoWriteBack_DoesNotWriteSecondary(t *testing.T) {
	ctx, _, secondary, m := setup(t)
	_ = m.Put(ctx, "key", "value")

	if _, err := secondary.Get(ctx, "key"); err == nil {
		t.Fatal("expected secondary to not have key without write-back")
	}
}

func TestPut_WithWriteBack_WritesSecondary(t *testing.T) {
	ctx := context.Background()
	primary := mock.New()
	secondary := mock.New()
	m, _ := mirror.New(primary, secondary, mirror.WithWriteBack())

	_ = m.Put(ctx, "api/key", "tok")

	val, err := secondary.Get(ctx, "api/key")
	if err != nil {
		t.Fatalf("secondary Get: %v", err)
	}
	if val != "tok" {
		t.Fatalf("want %q, got %q", "tok", val)
	}
}

func TestDelete_RemovesFromBoth(t *testing.T) {
	ctx, primary, secondary, m := setup(t)
	_ = primary.Put(ctx, "x", "1")
	_ = secondary.Put(ctx, "x", "1")

	if err := m.Delete(ctx, "x"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := primary.Get(ctx, "x"); err == nil {
		t.Fatal("expected primary key to be deleted")
	}
	if _, err := secondary.Get(ctx, "x"); err == nil {
		t.Fatal("expected secondary key to be deleted")
	}
}

func TestList_ReturnsPrimaryKeys(t *testing.T) {
	ctx, primary, _, m := setup(t)
	_ = primary.Put(ctx, "a", "1")
	_ = primary.Put(ctx, "b", "2")

	keys, err := m.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(keys))
	}
}
