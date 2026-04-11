package watch_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vaultshift/internal/diff"
	"github.com/vaultshift/internal/provider"
	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/watch"
)

func setupWatcher(t *testing.T, alerts *[]diff.Result, mu *sync.Mutex) (*watch.Watcher, *mock.Provider) {
	t.Helper()
	reg := provider.NewRegistry()
	mp := mock.New()
	reg.Register("mock", mp)
	w := watch.New(reg, 20*time.Millisecond, func(key string, r diff.Result) {
		mu.Lock()
		*alerts = append(*alerts, r)
		mu.Unlock()
	})
	return w, mp
}

func TestSnapshot_CapturesBaseline(t *testing.T) {
	var alerts []diff.Result
	var mu sync.Mutex
	w, mp := setupWatcher(t, &alerts, &mu)
	ctx := context.Background()
	_ = mp.PutSecret(ctx, "key1", "v1")
	if err := w.Snapshot(ctx, "mock", []string{"key1"}); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
}

func TestSnapshot_UnknownProvider(t *testing.T) {
	reg := provider.NewRegistry()
	w := watch.New(reg, time.Second, func(_ string, _ diff.Result) {})
	err := w.Snapshot(context.Background(), "missing", []string{"k"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestStart_DetectsDrift(t *testing.T) {
	var alerts []diff.Result
	var mu sync.Mutex
	w, mp := setupWatcher(t, &alerts, &mu)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = mp.PutSecret(ctx, "key1", "original")
	_ = w.Snapshot(ctx, "mock", []string{"key1"})

	_ = w.Start(ctx, "mock", []string{"key1"})
	// Mutate the secret after starting the watcher
	_ = mp.PutSecret(ctx, "key1", "changed")

	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		mu.Lock()
		count := len(alerts)
		mu.Unlock()
		if count > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(alerts) == 0 {
		t.Fatal("expected drift alert but got none")
	}
	if alerts[0].Status != diff.Updated {
		t.Fatalf("expected Updated status, got %v", alerts[0].Status)
	}
}

func TestStart_NoDriftWhenUnchanged(t *testing.T) {
	var alerts []diff.Result
	var mu sync.Mutex
	w, mp := setupWatcher(t, &alerts, &mu)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = mp.PutSecret(ctx, "stable", "same")
	_ = w.Snapshot(ctx, "mock", []string{"stable"})
	_ = w.Start(ctx, "mock", []string{"stable"})

	time.Sleep(80 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestStart_UnknownProvider(t *testing.T) {
	reg := provider.NewRegistry()
	w := watch.New(reg, time.Second, func(_ string, _ diff.Result) {})
	err := w.Start(context.Background(), "nope", []string{"k"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
