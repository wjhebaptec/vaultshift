package notify_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultshift/internal/notify"
)

func TestSend_NoHandlers(t *testing.T) {
	n := notify.New()
	if err := n.Send(notify.Event{Type: notify.EventRotated, Secret: "s"}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSend_SingleHandler(t *testing.T) {
	var received notify.Event
	h := func(e notify.Event) error { received = e; return nil }
	n := notify.New(h)
	evt := notify.Event{Type: notify.EventSynced, Secret: "db/pass", Provider: "gcp"}
	if err := n.Send(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Secret != "db/pass" {
		t.Errorf("expected secret db/pass, got %s", received.Secret)
	}
}

func TestSend_SetsTimestamp(t *testing.T) {
	var received notify.Event
	n := notify.New(func(e notify.Event) error { received = e; return nil })
	n.Send(notify.Event{Type: notify.EventDriftDet}) //nolint:errcheck
	if received.Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically")
	}
}

func TestSend_PreservesTimestamp(t *testing.T) {
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var received notify.Event
	n := notify.New(func(e notify.Event) error { received = e; return nil })
	n.Send(notify.Event{Type: notify.EventRotated, Timestamp: fixed}) //nolint:errcheck
	if !received.Timestamp.Equal(fixed) {
		t.Errorf("expected fixed timestamp, got %v", received.Timestamp)
	}
}

func TestSend_MultipleHandlers_AllCalled(t *testing.T) {
	count := 0
	h := func(e notify.Event) error { count++; return nil }
	n := notify.New(h, h, h)
	n.Send(notify.Event{Type: notify.EventFailed}) //nolint:errcheck
	if count != 3 {
		t.Errorf("expected 3 handler calls, got %d", count)
	}
}

func TestSend_HandlerError_ReturnsError(t *testing.T) {
	bad := func(e notify.Event) error { return errors.New("boom") }
	n := notify.New(bad)
	if err := n.Send(notify.Event{Type: notify.EventFailed}); err == nil {
		t.Error("expected error from failing handler")
	}
}

func TestLogHandler_WritesFormattedLine(t *testing.T) {
	var buf strings.Builder
	n := notify.New(notify.LogHandler(func(s string) { buf.WriteString(s) }))
	n.Send(notify.Event{Type: notify.EventRotated, Secret: "k", Provider: "aws", Message: "ok"}) //nolint:errcheck
	if !strings.Contains(buf.String(), "rotated") {
		t.Errorf("expected log line to contain 'rotated', got: %s", buf.String())
	}
}

func TestRegister_AddsHandler(t *testing.T) {
	count := 0
	n := notify.New()
	n.Register(func(e notify.Event) error { count++; return nil })
	n.Send(notify.Event{Type: notify.EventSynced}) //nolint:errcheck
	if count != 1 {
		t.Errorf("expected 1 call after Register, got %d", count)
	}
}
