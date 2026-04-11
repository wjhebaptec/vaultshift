// Package event provides a simple publish-subscribe event bus for
// broadcasting lifecycle events across vaultshift components.
package event

import (
	"sync"
	"time"
)

// Type represents a named event category.
type Type string

const (
	TypeRotated  Type = "secret.rotated"
	TypeSynced   Type = "secret.synced"
	TypeExpired  Type = "secret.expired"
	TypeAccessed Type = "secret.accessed"
	TypeError    Type = "operation.error"
)

// Event carries metadata about something that occurred in the system.
type Event struct {
	Type      Type
	Key       string
	Provider  string
	Message   string
	OccurredAt time.Time
	Meta      map[string]string
}

// Handler is a function that receives an event.
type Handler func(e Event)

// Bus is a simple in-process pub/sub event dispatcher.
type Bus struct {
	mu       sync.RWMutex
	handlers map[Type][]Handler
}

// New returns an initialised Bus.
func New() *Bus {
	return &Bus{handlers: make(map[Type][]Handler)}
}

// Subscribe registers a handler for the given event type.
func (b *Bus) Subscribe(t Type, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[t] = append(b.handlers[t], h)
}

// Publish dispatches an event to all registered handlers.
// If OccurredAt is zero it is set to the current UTC time.
func (b *Bus) Publish(e Event) {
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers[e.Type]))
	copy(handlers, b.handlers[e.Type])
	b.mu.RUnlock()

	for _, h := range handlers {
		h(e)
	}
}

// Unsubscribe removes all handlers for the given event type.
func (b *Bus) Unsubscribe(t Type) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, t)
}
