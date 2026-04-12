package observe_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/event"
	"github.com/vaultshift/internal/metrics"
	"github.com/vaultshift/internal/observe"
)

func setup(t *testing.T) (*observe.Observer, *metrics.Collector, *bytes.Buffer, *event.Bus) {
	t.Helper()
	collector := metrics.New()
	var buf bytes.Buffer
	logger := audit.New(&buf)
	bus := event.New()
	obs := observe.New("aws", collector, logger, bus)
	return obs, collector, &buf, bus
}

func TestRecord_SuccessUpdatesMetrics(t *testing.T) {
	obs, collector, _, _ := setup(t)
	obs.Record(context.Background(), observe.KindGet, "db/pass", nil)
	summary := collector.Summary()
	if summary[observe.KindGet] != 1 {
		t.Fatalf("expected 1 metric entry, got %d", summary[observe.KindGet])
	}
}

func TestRecord_ErrorStatusInMetrics(t *testing.T) {
	obs, collector, _, _ := setup(t)
	obs.Record(context.Background(), observe.KindPut, "db/pass", errors.New("boom"))
	entries := collector.Entries()
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	if entries[0].Status != "error" {
		t.Fatalf("expected status 'error', got %q", entries[0].Status)
	}
}

func TestRecord_WritesAuditLog(t *testing.T) {
	obs, _, buf, _ := setup(t)
	obs.Record(context.Background(), observe.KindDelete, "api/key", nil)
	line := buf.String()
	if !strings.Contains(line, "api/key") {
		t.Fatalf("audit log missing key: %s", line)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &m); err != nil {
		t.Fatalf("audit log is not valid JSON: %v", err)
	}
}

func TestRecord_PublishesEvent(t *testing.T) {
	obs, _, _, bus := setup(t)
	received := make(chan event.Event, 1)
	bus.Subscribe(observe.KindGet, func(e event.Event) {
		received <- e
	})
	obs.Record(context.Background(), observe.KindGet, "svc/token", nil)
	ev := <-received
	if ev.Kind != observe.KindGet {
		t.Fatalf("expected kind %q, got %q", observe.KindGet, ev.Kind)
	}
}

func TestRecord_ProviderTaggedInEvent(t *testing.T) {
	obs, _, _, bus := setup(t)
	received := make(chan event.Event, 1)
	bus.Subscribe(observe.KindPut, func(e event.Event) {
		received <- e
	})
	obs.Record(context.Background(), observe.KindPut, "cfg/url", nil)
	ev := <-received
	payload, ok := ev.Payload.(map[string]string)
	if !ok {
		t.Fatal("payload not map[string]string")
	}
	if payload["provider"] != "aws" {
		t.Fatalf("expected provider 'aws', got %q", payload["provider"])
	}
}
