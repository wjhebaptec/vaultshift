package mux_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vaultshift/internal/mux"
	"github.com/vaultshift/internal/provider/mock"
)

func routerByPrefix(key string) string {
	if strings.HasPrefix(key, "aws/") {
		return "aws"
	}
	return "gcp"
}

func setupMux(t *testing.T) (*mux.Mux, *mock.Provider, *mock.Provider) {
	t.Helper()
	m, err := mux.New(routerByPrefix)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	aws := mock.New()
	gcp := mock.New()
	_ = m.Register("aws", aws)
	_ = m.Register("gcp", gcp)
	return m, aws, gcp
}

func TestNew_NilRouter_ReturnsError(t *testing.T) {
	_, err := mux.New(nil)
	if err == nil {
		t.Fatal("expected error for nil router")
	}
}

func TestRegister_EmptyName_ReturnsError(t *testing.T) {
	m, _ := mux.New(routerByPrefix)
	if err := m.Register("", mock.New()); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegister_NilProvider_ReturnsError(t *testing.T) {
	m, _ := mux.New(routerByPrefix)
	if err := m.Register("aws", nil); err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestPutAndGet_RoutedByPrefix(t *testing.T) {
	m, aws, gcp := setupMux(t)
	ctx := context.Background()

	if err := m.Put(ctx, "aws/db-pass", "secret1"); err != nil {
		t.Fatalf("Put aws: %v", err)
	}
	if err := m.Put(ctx, "gcp/api-key", "secret2"); err != nil {
		t.Fatalf("Put gcp: %v", err)
	}

	if v, _ := aws.Get(ctx, "aws/db-pass"); v != "secret1" {
		t.Errorf("aws store: got %q, want %q", v, "secret1")
	}
	if v, _ := gcp.Get(ctx, "gcp/api-key"); v != "secret2" {
		t.Errorf("gcp store: got %q, want %q", v, "secret2")
	}
}

func TestGet_UnknownRoute_ReturnsError(t *testing.T) {
	m, _, _ := setupMux(t)
	// router maps non-aws keys to "gcp"; unregister gcp by using a fresh mux
	m2, _ := mux.New(func(key string) string { return "unknown" })
	_ = m2.Register("aws", mock.New())
	_, err := m2.Get(context.Background(), "some/key")
	if err == nil {
		t.Fatal("expected error for unregistered route")
	}
}

func TestDelete_RoutedCorrectly(t *testing.T) {
	m, aws, _ := setupMux(t)
	ctx := context.Background()
	_ = m.Put(ctx, "aws/token", "tok")
	if err := m.Delete(ctx, "aws/token"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	keys, _ := aws.List(ctx)
	for _, k := range keys {
		if k == "aws/token" {
			t.Error("key should have been deleted")
		}
	}
}

func TestList_AggregatesAllProviders(t *testing.T) {
	m, _, _ := setupMux(t)
	ctx := context.Background()
	_ = m.Put(ctx, "aws/a", "1")
	_ = m.Put(ctx, "gcp/b", "2")
	keys, err := m.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("got %d keys, want 2", len(keys))
	}
}
