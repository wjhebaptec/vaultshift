package resolve_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/resolve"
)

func TestChain_GetFromFirst(t *testing.T) {
	c := resolve.NewChain()
	c.Add("primary", &stubProvider{data: map[string]string{"k": "v1"}})
	c.Add("secondary", &stubProvider{data: map[string]string{"k": "v2"}})

	val, src, err := c.Get(context.Background(), "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "v1" {
		t.Errorf("expected v1, got %s", val)
	}
	if src != "primary" {
		t.Errorf("expected source primary, got %s", src)
	}
}

func TestChain_FallsBackToSecond(t *testing.T) {
	c := resolve.NewChain()
	c.Add("primary", &stubProvider{data: map[string]string{}})
	c.Add("secondary", &stubProvider{data: map[string]string{"k": "v2"}})

	val, src, err := c.Get(context.Background(), "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "v2" {
		t.Errorf("expected v2, got %s", val)
	}
	if src != "secondary" {
		t.Errorf("expected source secondary, got %s", src)
	}
}

func TestChain_NotFound(t *testing.T) {
	c := resolve.NewChain()
	c.Add("a", &stubProvider{data: map[string]string{}})
	c.Add("b", &stubProvider{data: map[string]string{}})

	_, _, err := c.Get(context.Background(), "missing")
	if !errors.Is(err, resolve.ErrNotResolved) {
		t.Errorf("expected ErrNotResolved, got %v", err)
	}
}

func TestChain_Providers_ReturnedInOrder(t *testing.T) {
	p1 := &stubProvider{data: map[string]string{}}
	p2 := &stubProvider{data: map[string]string{}}
	c := resolve.NewChain()
	c.Add("first", p1)
	c.Add("second", p2)

	providers := c.Providers()
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
}

func TestChain_EmptyChain_ReturnsError(t *testing.T) {
	c := resolve.NewChain()
	_, _, err := c.Get(context.Background(), "any")
	if !errors.Is(err, resolve.ErrNotResolved) {
		t.Errorf("expected ErrNotResolved on empty chain, got %v", err)
	}
}
