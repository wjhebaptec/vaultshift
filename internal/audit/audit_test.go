package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vaultshift/internal/audit"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	l := audit.New(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLog_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	err := l.Log(audit.EventRotate, "aws", "prod/db/password", true, "rotated successfully")
	if err != nil {
		t.Fatalf("Log() error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var e audit.Event
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if e.Type != audit.EventRotate {
		t.Errorf("expected type %q, got %q", audit.EventRotate, e.Type)
	}
	if e.Provider != "aws" {
		t.Errorf("expected provider %q, got %q", "aws", e.Provider)
	}
	if e.SecretKey != "prod/db/password" {
		t.Errorf("expected secret_key %q, got %q", "prod/db/password", e.SecretKey)
	}
	if !e.Success {
		t.Error("expected success=true")
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLog_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	events := []struct {
		typ     audit.EventType
		provider string
		key     string
	}{
		{audit.EventSync, "gcp", "secret/foo"},
		{audit.EventDelete, "vault", "secret/bar"},
		{audit.EventAccess, "aws", "secret/baz"},
	}

	for _, ev := range events {
		if err := l.Log(ev.typ, ev.provider, ev.key, true, ""); err != nil {
			t.Fatalf("Log() error: %v", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 log lines, got %d", len(lines))
	}
}

func TestLogEvent_SetsTimestampIfZero(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	e := audit.Event{
		Type:      audit.EventSync,
		Provider:  "vault",
		SecretKey: "kv/myapp/token",
		Success:   false,
		Message:   "provider unreachable",
	}
	if err := l.LogEvent(e); err != nil {
		t.Fatalf("LogEvent() error: %v", err)
	}

	var got audit.Event
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got.Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically")
	}
	if got.Timestamp.After(time.Now().UTC().Add(time.Second)) {
		t.Error("timestamp is unexpectedly in the future")
	}
}
