package debounce_test

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/vaultshift/internal/debounce"
)

func TestNew_InvalidWindow_ReturnsError(t *testing.T) {
	_, err := debounce.New(0, func(_ context.Context, _ string) {})
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNew_NilFunc_ReturnsError(t *testing.T) {
	_, err := debounce.New(10*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for nil fn")
	}
}

func TestTrigger_CallsFnAfterWindow(t *testing.T) {
	var mu sync.Mutex
	called := []string{}

	d, err := debounce.New(30*time.Millisecond, func(_ context.Context, key string) {
		mu.Lock()
		called = append(called, key)
		mu.Unlock()
	})
	if err != nil {
		t.Fatal(err)
	}

	d.Trigger(context.Background(), "db/password")
	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(called) != 1 || called[0] != "db/password" {
		t.Fatalf("expected [db/password], got %v", called)
	}
}

func TestTrigger_ResetsTimerOnRepeat(t *testing.T) {
	var mu sync.Mutex
	count := 0

	d, _ := debounce.New(40*time.Millisecond, func(_ context.Context, _ string) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	for i := 0; i < 5; i++ {
		d.Trigger(context.Background(), "api/key")
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Fatalf("expected fn called once, got %d", count)
	}
}

func TestCancel_StopsPendingCall(t *testing.T) {
	var mu sync.Mutex
	called := false

	d, _ := debounce.New(50*time.Millisecond, func(_ context.Context, _ string) {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	d.Trigger(context.Background(), "token")
	d.Cancel("token")
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if called {
		t.Fatal("expected fn not to be called after cancel")
	}
}

func TestPending_ReturnsActiveKeys(t *testing.T) {
	d, _ := debounce.New(200*time.Millisecond, func(_ context.Context, _ string) {})

	d.Trigger(context.Background(), "alpha")
	d.Trigger(context.Background(), "beta")

	pending := d.Pending()
	sort.Strings(pending)

	if len(pending) != 2 || pending[0] != "alpha" || pending[1] != "beta" {
		t.Fatalf("unexpected pending keys: %v", pending)
	}

	d.Cancel("alpha")
	d.Cancel("beta")
}
