// Package audit provides structured logging of secret rotation and sync events.
package audit

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// EventType represents the kind of audit event.
type EventType string

const (
	EventRotate EventType = "rotate"
	EventSync   EventType = "sync"
	EventDelete EventType = "delete"
	EventAccess EventType = "access"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	Provider  string    `json:"provider"`
	SecretKey string    `json:"secret_key"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes audit events to an io.Writer as newline-delimited JSON.
type Logger struct {
	w       io.Writer
	encoder *json.Encoder
}

// New creates a new audit Logger writing to w.
// Pass nil to write to os.Stdout.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{
		w:       w,
		encoder: json.NewEncoder(w),
	}
}

// Log records an audit event.
func (l *Logger) Log(eventType EventType, provider, secretKey string, success bool, message string) error {
	e := Event{
		Timestamp: time.Now().UTC(),
		Type:      eventType,
		Provider:  provider,
		SecretKey: secretKey,
		Success:   success,
		Message:   message,
	}
	return l.encoder.Encode(e)
}

// LogEvent records a pre-built Event.
func (l *Logger) LogEvent(e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	return l.encoder.Encode(e)
}
