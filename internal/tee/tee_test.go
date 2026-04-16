package tee_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/tee"
)

func setup(t *testing.T) (context.Context, *mock.Provider, *mock.Provider, *tee.Tee) {
	t.Helper()
	primary := mock.New()
	secondary := mock.New()
	tw, err := tee.New(primary, secondary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return context.Background(), primary, secondary, tw
}

func TestNew_NilPrimary_ReturnsError(t *testing.T) {
	_, err := tee.New(nil, mock.New())
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNew_NilSecondary_ReturnsError(t *testing.T) {
	_, err := tee.New(mock.New(), nil)
	if err == nil {
		t.Fatal("expected error for nil secondary")
	}
}

func TestGet_MirrorsToSecondary(t *testing.T) {
	ctx, primary, secondary, tw := setup(t)
	_ = primary.Put(ctx, "key", "val")
	got, err := tw.Get(ctx, "key")
	if err != nil || got != "val" {
		t.Fatalf("Get: got %q, err %v", got, err)
	}
	v, err := secondary.Get(ctx, "key")
	if err != nil || v != "val" {
		t.Fatalf("secondary not mirrored: got %q, err %v", v, err)
	}
}

func TestPut_MirrorsToBoth(t *testing.T) {
	ctx, primary, secondary, tw := setup(t)
	if err := tw.Put(ctx, "k", "v"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	for _, p := range []interface{ Get(context.Context, string) (string, error) }{primary, secondary} {
		v, err := p.Get(ctx, "k")
		if err != nil || v != "v" {
			t.Fatalf("expected v, got %q err %v", v, err)
		}
	}
}

func TestDelete_MirrorsToBoth(t *testing.T) {
	ctx, primary, secondary, tw := setup(t)
	_ = primary.Put(ctx, "k", "v")
	_ = secondary.Put(ctx, "k", "v")
	if err := tw.Delete(ctx, "k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := primary.Get(ctx, "k"); err == nil {
		t.Fatal("expected primary key removed")
	}
	if _, err := secondary.Get(ctx, "k"); err == nil {
		t.Fatal("expected secondary key removed")
	}
}

func TestWriteOnly_GetDoesNotMirror(t *testing.T) {
	ctx := context.Background()
	primary := mock.New()
	secondary := mock.New()
	tw, _ := tee.New(primary, secondary, tee.WithWriteOnly())
	_ = primary.Put(ctx, "k", "v")
	_, _ = tw.Get(ctx, "k")
	if _, err := secondary.Get(ctx, "k"); err == nil {
		t.Fatal("write-only: secondary should not have key after Get")
	}
}
