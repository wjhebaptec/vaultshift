package consensus_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/consensus"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
)

func setupReader(t *testing.T, minAgree int, values map[string]string) (*consensus.Reader, []string) {
	t.Helper()
	reg := provider.NewRegistry()
	names := []string{"p1", "p2", "p3"}
	for _, name := range names {
		p := mock.New()
		for k, v := range values {
			_ = p.PutSecret(context.Background(), k, v)
		}
		reg.Register(name, p)
	}
	r, err := consensus.New(reg, names, minAgree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return r, names
}

func TestGet_AllAgree(t *testing.T) {
	r, _ := setupReader(t, 2, map[string]string{"db/pass": "s3cr3t"})
	val, err := r.Get(context.Background(), "db/pass")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %s", val)
	}
}

func TestGet_QuorumMet_WithOneDisagreement(t *testing.T) {
	reg := provider.NewRegistry()
	p1 := mock.New()
	p2 := mock.New()
	p3 := mock.New()
	_ = p1.PutSecret(context.Background(), "key", "alpha")
	_ = p2.PutSecret(context.Background(), "key", "alpha")
	_ = p3.PutSecret(context.Background(), "key", "beta")
	reg.Register("p1", p1)
	reg.Register("p2", p2)
	reg.Register("p3", p3)

	r, _ := consensus.New(reg, []string{"p1", "p2", "p3"}, 2)
	val, err := r.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "alpha" {
		t.Fatalf("expected alpha, got %s", val)
	}
}

func TestGet_NoConsensus(t *testing.T) {
	reg := provider.NewRegistry()
	for i, v := range []string{"x", "y", "z"} {
		p := mock.New()
		_ = p.PutSecret(context.Background(), "key", v)
		reg.Register(fmt.Sprintf("p%d", i), p)
	}
	r, _ := consensus.New(reg, []string{"p0", "p1", "p2"}, 2)
	_, err := r.Get(context.Background(), "key")
	if err != consensus.ErrNoConsensus {
		t.Fatalf("expected ErrNoConsensus, got %v", err)
	}
}

func TestGet_InsufficientProviders(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register("p1", mock.New()) // no secret stored
	r, _ := consensus.New(reg, []string{"p1"}, 1)
	_, err := r.Get(context.Background(), "missing")
	if err != consensus.ErrInsufficientProviders {
		t.Fatalf("expected ErrInsufficientProviders, got %v", err)
	}
}

func TestNew_InvalidMinAgree(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register("p1", mock.New())
	_, err := consensus.New(reg, []string{"p1"}, 0)
	if err == nil {
		t.Fatal("expected error for minAgree=0")
	}
}

func TestGet_EmptyKey_ReturnsError(t *testing.T) {
	r, _ := setupReader(t, 1, nil)
	_, err := r.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestProviders_ReturnsCopy(t *testing.T) {
	r, names := setupReader(t, 1, nil)
	got := r.Providers()
	if len(got) != len(names) {
		t.Fatalf("expected %d providers, got %d", len(names), len(got))
	}
}
