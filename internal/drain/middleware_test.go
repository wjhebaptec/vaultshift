package drain_test

import (
	"context"
	"testing"
	"time"

	"github.com/vaultshift/internal/drain"
	"github.com/vaultshift/internal/provider/mock"
)

func newWrapped(t *testing.T) (*drain.WrappedProvider, *drain.Drainer) {
	t.Helper()
	m := mock.New("test")
	d := drain.New(drain.WithTimeout(500 * time.Millisecond))
	return drain.Wrap(m, d), d
}

func TestWrapGet_DelegatesToInner(t *testing.T) {
	w, _ := newWrapped(t)
	ctx := context.Background()
	_ = w.PutSecret(ctx, "k", "v")
	val, err := w.GetSecret(ctx, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "v" {
		t.Fatalf("expected v, got %s", val)
	}
}

func TestWrapPut_DelegatesToInner(t *testing.T) {
	w, _ := newWrapped(t)
	if err := w.PutSecret(context.Background(), "key", "val"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWrapDelete_DelegatesToInner(t *testing.T) {
	w, _ := newWrapped(t)
	ctx := context.Background()
	_ = w.PutSecret(ctx, "del", "x")
	if err := w.DeleteSecret(ctx, "del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWrapList_DelegatesToInner(t *testing.T) {
	w, _ := newWrapped(t)
	ctx := context.Background()
	_ = w.PutSecret(ctx, "a", "1")
	keys, err := w.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected at least one key")
	}
}

func TestWrapGet_FailsWhenDrained(t *testing.T) {
	w, d := newWrapped(t)
	// drain without any in-flight ops
	go func() { _ = d.Drain(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	_, err := w.GetSecret(context.Background(), "k")
	if err == nil {
		t.Fatal("expected error after drain, got nil")
	}
}

func TestWrapPut_FailsWhenDrained(t *testing.T) {
	w, d := newWrapped(t)
	// drain without any in-flight ops
	go func() { _ = d.Drain(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	err := w.PutSecret(context.Background(), "k", "v")
	if err == nil {
		t.Fatal("expected error after drain, got nil")
	}
}

func TestWrapName_ReturnsInnerName(t *testing.T) {
	w, _ := newWrapped(t)
	if w.Name() != "test" {
		t.Fatalf("expected 'test', got %s", w.Name())
	}
}
