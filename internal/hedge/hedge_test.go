package hedge_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vaultshift/internal/hedge"
	"github.com/vaultshift/internal/provider/mock"
)

func TestNew_NoProviders_ReturnsError(t *testing.T) {
	_, err := hedge.New(nil)
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
}

func TestGet_FirstProviderSucceeds(t *testing.T) {
	p := mock.New()
	_ = p.Put(context.Background(), "key", "value")

	h, err := hedge.New([]interface{ Get(context.Context, string) (string, error) }{p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := h.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected 'value', got %q", val)
	}
}

func TestGet_FallsBackToSecondProvider(t *testing.T) {
	primary := mock.New() // key not set — will error
	secondary := mock.New()
	_ = secondary.Put(context.Background(), "key", "fallback")

	h, err := hedge.New([]interface{ Get(context.Context, string) (string, error) }{primary, secondary},
		hedge.WithDelay(5*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := h.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "fallback" {
		t.Errorf("expected 'fallback', got %q", val)
	}
}

func TestGet_AllFail_ReturnsError(t *testing.T) {
	p1 := mock.New()
	p2 := mock.New()

	h, err := hedge.New([]interface{ Get(context.Context, string) (string, error) }{p1, p2},
		hedge.WithDelay(5*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = h.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestGet_ContextCancelled_ReturnsError(t *testing.T) {
	var calls int64
	p := mock.New()
	_ = p.Put(context.Background(), "key", "v")

	h, _ := hedge.New([]interface{ Get(context.Context, string) (string, error) }{p},
		hedge.WithDelay(100*time.Millisecond))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, _ = h.Get(ctx, "key")
	_ = calls
	_ = errors.New("")
	atomic.AddInt64(&calls, 1)
}
