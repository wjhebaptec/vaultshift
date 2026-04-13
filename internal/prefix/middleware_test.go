package prefix

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
)

func newWrapped(t *testing.T, ns string) (*Wrapped, *mock.Provider) {
	t.Helper()
	m := mock.New()
	pfx, err := New(ns)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	w, err := Wrap(m, pfx)
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	return w, m
}

func TestWrap_PutAndGet_PrefixTransparent(t *testing.T) {
	w, _ := newWrapped(t, "prod")
	ctx := context.Background()
	if err := w.Put(ctx, "db/pass", "secret"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	val, err := w.Get(ctx, "db/pass")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "secret" {
		t.Errorf("got %q, want %q", val, "secret")
	}
}

func TestWrap_Delete_RemovesPrefixedKey(t *testing.T) {
	w, _ := newWrapped(t, "prod")
	ctx := context.Background()
	_ = w.Put(ctx, "token", "abc")
	if err := w.Delete(ctx, "token"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := w.Get(ctx, "token")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestWrap_List_StripsPrefix(t *testing.T) {
	w, _ := newWrapped(t, "ns")
	ctx := context.Background()
	_ = w.Put(ctx, "alpha", "1")
	_ = w.Put(ctx, "beta", "2")
	keys, err := w.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(keys))
	}
	for _, k := range keys {
		if k != "alpha" && k != "beta" {
			t.Errorf("unexpected key %q", k)
		}
	}
}

func TestWrap_NilProvider_ReturnsError(t *testing.T) {
	pfx, _ := New("ns")
	_, err := Wrap(nil, pfx)
	if err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestWrap_NilPrefixer_ReturnsError(t *testing.T) {
	m := mock.New()
	_, err := Wrap(m, nil)
	if err == nil {
		t.Error("expected error for nil prefixer")
	}
}
