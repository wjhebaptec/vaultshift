// Package notify provides event notification hooks for secret rotation and sync operations.
package notify

import (
	"fmt"
	"time"
)

// EventType classifies the kind of notification event.
type EventType string

const (
	EventRotated  EventType = "rotated"
	EventSynced   EventType = "synced"
	EventFailed   EventType = "failed"
	EventDriftDet EventType = "drift_detected"
)

// Event holds metadata about a notification.
type Event struct {
	Type      EventType
	Secret    string
	Provider  string
	Message   string
	Timestamp time.Time
}

// Handler is a function that receives a notification event.
type Handler func(Event) error

// Notifier dispatches events to one or more registered handlers.
type Notifier struct {
	handlers []Handler
}

// New creates a Notifier with the given handlers.
func New(handlers ...Handler) *Notifier {
	return &Notifier{handlers: handlers}
}

// Register adds a new handler to the notifier.
func (n *Notifier) Register(h Handler) {
	n.handlers = append(n.handlers, h)
}

// Send dispatches an event to all registered handlers, collecting errors.
func (n *Notifier) Send(evt Event) error {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now().UTC()
	}
	var errs []error
	for _, h := range n.handlers {
		if err := h(evt); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notify: %d handler(s) failed, first: %w", len(errs), errs[0])
	}
	return nil
}

// LogHandler returns a Handler that writes events to any fmt.Stringer-compatible sink via a format func.
func LogHandler(logFn func(string)) Handler {
	return func(evt Event) error {
		logFn(fmt.Sprintf("[%s] type=%s secret=%s provider=%s msg=%s",
			evt.Timestamp.Format(time.RFC3339), evt.Type, evt.Secret, evt.Provider, evt.Message))
		return nil
	}
}
