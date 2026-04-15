package signal_test

import (
	"testing"

	"github.com/vaultshift/internal/signal"
)

func TestOn_EmptyName_ReturnsError(t *testing.T) {
	b := signal.New()
	err := b.On("", func(_ string, _ any) {})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestOn_NilHandler_ReturnsError(t *testing.T) {
	b := signal.New()
	err := b.On("x", nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestEmit_CallsRegisteredHandler(t *testing.T) {
	b := signal.New()
	called := false
	_ = b.On("ev", func(name string, payload any) {
		called = true
		if name != "ev" {
			t.Errorf("unexpected name %q", name)
		}
		if payload != "hello" {
			t.Errorf("unexpected payload %v", payload)
		}
	})
	b.Emit("ev", "hello")
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestEmit_NoHandlers_DoesNotPanic(t *testing.T) {
	b := signal.New()
	b.Emit("unknown", nil) // must not panic
}

func TestEmit_MultipleHandlers_AllCalled(t *testing.T) {
	b := signal.New()
	count := 0
	for i := 0; i < 3; i++ {
		_ = b.On("tick", func(_ string, _ any) { count++ })
	}
	b.Emit("tick", nil)
	if count != 3 {
		t.Fatalf("expected 3 calls, got %d", count)
	}
}

func TestOff_RemovesHandlers(t *testing.T) {
	b := signal.New()
	called := false
	_ = b.On("ev", func(_ string, _ any) { called = true })
	b.Off("ev")
	b.Emit("ev", nil)
	if called {
		t.Fatal("handler should not have been called after Off")
	}
}

func TestNames_ReturnsRegisteredSignals(t *testing.T) {
	b := signal.New()
	_ = b.On("a", func(_ string, _ any) {})
	_ = b.On("b", func(_ string, _ any) {})
	names := b.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}
