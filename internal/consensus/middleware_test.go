package consensus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/vaultshift/internal/consensus"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupGuarded(t *testing.T) (*consensus.GuardedProvider, *mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	primary := mock.New()
	names := []string{"p1", "p2", "p3"}
	for _, name := range names {
		p := mock.New()
		_ = p.PutSecret(context.Background(), "token", "abc123")
		reg.Register(name, p)
	}
	r, _ := consensus.New(reg, names, 2)
	g, err := consensus.NewGuardedProvider(primary, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return g, primary
}

func TestGuardedProvider_GetUsesConsensus(t *testing.T) {
	g, _ := setupGuarded(t)
	val, err := g.GetSecret(context.Background(), "token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "abc123" {
		t.Fatalf("expected abc123, got %s", val)
	}
}

func TestGuardedProvider_PutWritesToPrimary(t *testing.T) {
	g, primary := setupGuarded(t)
	err := g.PutSecret(context.Background(), "newkey", "newval")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := primary.GetSecret(context.Background(), "newkey")
	if err != nil {
		t.Fatalf("expected value in primary, got error: %v", err)
	}
	if val != "newval" {
		t.Fatalf("expected newval, got %s", val)
	}
}

func TestGuardedProvider_DeleteGoesToPrimary(t *testing.T) {
	g, primary := setupGuarded(t)
	_ = primary.PutSecret(context.Background(), "tmp", "v")
	if err := g.DeleteSecret(context.Background(), "tmp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := primary.GetSecret(context.Background(), "tmp")
	if err == nil {
		t.Fatal("expected key to be deleted from primary")
	}
}

func TestGuardedProvider_ListGoesToPrimary(t *testing.T) {
	g, primary := setupGuarded(t)
	for i := 0; i < 3; i++ {
		_ = primary.PutSecret(context.Background(), fmt.Sprintf("k%d", i), "v")
	}
	keys, err := g.ListSecrets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
}

func TestNewGuardedProvider_NilPrimary_ReturnsError(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register("p1", mock.New())
	r, _ := consensus.New(reg, []string{"p1"}, 1)
	_, err := consensus.NewGuardedProvider(nil, r)
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNewGuardedProvider_NilReader_ReturnsError(t *testing.T) {
	_, err := consensus.NewGuardedProvider(mock.New(), nil)
	if err == nil {
		t.Fatal("expected error for nil reader")
	}
}
