package elect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultshift/internal/provider/mock"
)

func setupGuard(t *testing.T, candidate string) (*GuardedProvider, *Elector, *mock.Provider) {
	t.Helper()
	m := mock.New()
	e, _ := New(WithTTL(30 * time.Second))
	g, err := Guard(m, e, candidate)
	if err != nil {
		t.Fatalf("Guard: %v", err)
	}
	return g, e, m
}

func TestGuard_Put_AllowedWhenLeader(t *testing.T) {
	g, e, _ := setupGuard(t, "node-1")
	_ = e.Campaign("node-1")
	if err := g.Put(context.Background(), "k", "v"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGuard_Put_BlockedWhenNotLeader(t *testing.T) {
	g, e, _ := setupGuard(t, "node-1")
	_ = e.Campaign("node-2") // someone else is leader
	if err := g.Put(context.Background(), "k", "v"); !errors.Is(err, ErrNotLeader) {
		t.Fatalf("expected ErrNotLeader, got %v", err)
	}
}

func TestGuard_Get_AllowedWithoutLeadership(t *testing.T) {
	g, e, m := setupGuard(t, "node-1")
	_ = e.Campaign("node-2")
	_ = m.Put(context.Background(), "k", "secret")
	v, err := g.Get(context.Background(), "k")
	if err != nil || v != "secret" {
		t.Fatalf("Get failed: %v %q", err, v)
	}
}

func TestGuard_Delete_BlockedWhenNotLeader(t *testing.T) {
	g, e, _ := setupGuard(t, "node-1")
	_ = e.Campaign("node-2")
	if err := g.Delete(context.Background(), "k"); !errors.Is(err, ErrNotLeader) {
		t.Fatalf("expected ErrNotLeader, got %v", err)
	}
}

func TestGuard_List_AllowedWithoutLeadership(t *testing.T) {
	g, e, _ := setupGuard(t, "node-1")
	_ = e.Campaign("node-2")
	if _, err := g.List(context.Background()); err != nil {
		t.Fatalf("List should not require leadership: %v", err)
	}
}

func TestGuard_NilProvider_ReturnsError(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	if _, err := Guard(nil, e, "node-1"); err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestGuard_EmptyCandidate_ReturnsError(t *testing.T) {
	e, _ := New(WithTTL(10 * time.Second))
	if _, err := Guard(mock.New(), e, ""); err == nil {
		t.Fatal("expected error for empty candidate")
	}
}
