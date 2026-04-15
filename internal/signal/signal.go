// Package signal provides a lightweight pub/sub signal bus for broadcasting
// named events to registered listeners within vaultshift.
package signal

import (
	"fmt"
	"sync"
)

// Handler is a function invoked when a named signal is fired.
type Handler func(name string, payload any)

// Bus routes named signals to registered handlers.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New creates an empty Bus.
func New() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// On registers a handler for the given signal name.
// Returns an error if name is empty or handler is nil.
func (b *Bus) On(name string, h Handler) error {
	if name == "" {
		return fmt.Errorf("signal: name must not be empty")
	}
	if h == nil {
		return fmt.Errorf("signal: handler must not be nil")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[name] = append(b.handlers[name], h)
	return nil
}

// Emit fires all handlers registered for name, passing payload to each.
// Handlers are called synchronously in registration order.
func (b *Bus) Emit(name string, payload any) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers[name]))
	copy(handlers, b.handlers[name])
	b.mu.RUnlock()
	for _, h := range handlers {
		h(name, payload)
	}
}

// Off removes all handlers registered for the given signal name.
func (b *Bus) Off(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, name)
}

// Names returns a snapshot of all signal names that have at least one handler.
func (b *Bus) Names() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0, len(b.handlers))
	for k := range b.handlers {
		out = append(out, k)
	}
	return out
}
