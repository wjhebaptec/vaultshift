package event_test

import (
	"sync"
	"testing"
	"time"

	"github.com/vaultshift/internal/event"
)

func TestPublish_DeliversToSubscriber(t *testing.T) {
	bus := event.New()
	var got event.Event
	bus.Subscribe(event.TypeRotated, func(e event.Event) { got = e })

	bus.Publish(event.Event{Type: event.TypeRotated, Key: "my/key", Provider: "aws"})

	if got.Key != "my/key" {
		t.Fatalf("expected key my/key, got %s", got.Key)
	}
	if got.Provider != "aws" {
		t.Fatalf("expected provider aws, got %s", got.Provider)
	}
}

func TestPublish_SetsTimestampIfZero(t *testing.T) {
	bus := event.New()
	var got event.Event
	bus.Subscribe(event.TypeSynced, func(e event.Event) { got = e })

	before := time.Now().UTC()
	bus.Publish(event.Event{Type: event.TypeSynced})

	if got.OccurredAt.Before(before) {
		t.Fatal("expected OccurredAt to be set to now")
	}
}

func TestPublish_PreservesExistingTimestamp(t *testing.T) {
	bus := event.New()
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var got event.Event
	bus.Subscribe(event.TypeExpired, func(e event.Event) { got = e })

	bus.Publish(event.Event{Type: event.TypeExpired, OccurredAt: fixed})

	if !got.OccurredAt.Equal(fixed) {
		t.Fatalf("expected timestamp %v, got %v", fixed, got.OccurredAt)
	}
}

func TestPublish_MultipleHandlers_AllCalled(t *testing.T) {
	bus := event.New()
	var mu sync.Mutex
	var calls []string

	for _, id := range []string{"a", "b", "c"} {
		id := id
		bus.Subscribe(event.TypeError, func(e event.Event) {
			mu.Lock()
			calls = append(calls, id)
			mu.Unlock()
		})
	}

	bus.Publish(event.Event{Type: event.TypeError})

	if len(calls) != 3 {
		t.Fatalf("expected 3 handlers called, got %d", len(calls))
	}
}

func TestPublish_NoHandlers_DoesNotPanic(t *testing.T) {
	bus := event.New()
	bus.Publish(event.Event{Type: event.TypeAccessed, Key: "x"})
}

func TestUnsubscribe_RemovesHandlers(t *testing.T) {
	bus := event.New()
	called := false
	bus.Subscribe(event.TypeRotated, func(e event.Event) { called = true })
	bus.Unsubscribe(event.TypeRotated)

	bus.Publish(event.Event{Type: event.TypeRotated})

	if called {
		t.Fatal("handler should not have been called after Unsubscribe")
	}
}

func TestPublish_DifferentTypes_DoNotCross(t *testing.T) {
	bus := event.New()
	var rotatedCalled, syncedCalled bool

	bus.Subscribe(event.TypeRotated, func(e event.Event) { rotatedCalled = true })
	bus.Subscribe(event.TypeSynced, func(e event.Event) { syncedCalled = true })

	bus.Publish(event.Event{Type: event.TypeRotated})

	if !rotatedCalled {
		t.Fatal("rotated handler should have been called")
	}
	if syncedCalled {
		t.Fatal("synced handler should not have been called")
	}
}
