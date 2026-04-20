package jitter_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/vaultshift/internal/jitter"
)

// --- minimal in-memory provider ---

type memProvider struct {
	mu   sync.Mutex
	data map[string]string
}

func newMem() *memProvider { return &memProvider{data: map[string]string{}} }

func (m *memProvider) Get(_ context.Context, k string) (string, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	v, ok := m.data[k]
	if !ok { return "", errors.New("not found") }
	return v, nil
}
func (m *memProvider) Put(_ context.Context, k, v string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	m.data[k] = v; return nil
}
func (m *memProvider) Delete(_ context.Context, k string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	delete(m.data, k); return nil
}
func (m *memProvider) List(_ context.Context) ([]string, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	keys := make([]string, 0, len(m.data))
	for k := range m.data { keys = append(keys, k) }
	return keys, nil
}

// noSleep replaces the real sleep with a no-op for deterministic tests.
func noSleep(_ context.Context, _ time.Duration) error { return nil }

func TestNew_NilProvider_ReturnsError(t *testing.T) {
	_, err := jitter.New(nil, time.Millisecond)
	if err == nil { t.Fatal("expected error for nil provider") }
}

func TestNew_NegativeMax_ReturnsError(t *testing.T) {
	_, err := jitter.New(newMem(), -1)
	if err == nil { t.Fatal("expected error for non-positive max") }
}

func TestGet_NoJitter_Delegates(t *testing.T) {
	m := newMem()
	_ = m.Put(context.Background(), "k", "v")
	j, _ := jitter.New(m, time.Millisecond, jitter.WithSleepFunc(noSleep))
	got, err := j.Get(context.Background(), "k")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got != "v" { t.Fatalf("want v, got %s", got) }
}

func TestPut_WritesValue(t *testing.T) {
	m := newMem()
	j, _ := jitter.New(m, time.Millisecond, jitter.WithSleepFunc(noSleep))
	if err := j.Put(context.Background(), "key", "secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := m.Get(context.Background(), "key")
	if got != "secret" { t.Fatalf("want secret, got %s", got) }
}

func TestDelete_RemovesValue(t *testing.T) {
	m := newMem()
	_ = m.Put(context.Background(), "key", "val")
	j, _ := jitter.New(m, time.Millisecond, jitter.WithSleepFunc(noSleep))
	if err := j.Delete(context.Background(), "key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys, _ := m.List(context.Background())
	if len(keys) != 0 { t.Fatalf("expected empty store, got %v", keys) }
}

func TestPut_ContextCancelled_ReturnsError(t *testing.T) {
	blocking := func(ctx context.Context, _ time.Duration) error {
		<-ctx.Done(); return ctx.Err()
	}
	j, _ := jitter.New(newMem(), time.Hour, jitter.WithSleepFunc(blocking))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := j.Put(ctx, "k", "v"); !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

func TestList_NoJitter_Delegates(t *testing.T) {
	m := newMem()
	_ = m.Put(context.Background(), "a", "1")
	_ = m.Put(context.Background(), "b", "2")
	j, _ := jitter.New(m, time.Millisecond, jitter.WithSleepFunc(noSleep))
	keys, err := j.List(context.Background())
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(keys) != 2 { t.Fatalf("want 2 keys, got %d", len(keys)) }
}
