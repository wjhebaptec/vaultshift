package metrics_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/vaultshift/internal/metrics"
)

func TestPrintTable_ContainsHeaders(t *testing.T) {
	c := metrics.New()
	var buf bytes.Buffer
	r := metrics.NewReporter(c, &buf)
	r.PrintTable()
	out := buf.String()
	for _, want := range []string{"EVENT TYPE", "COUNT", "rotation", "sync", "error"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\noutput: %s", want, out)
		}
	}
}

func TestPrintTable_ShowsCounts(t *testing.T) {
	c := metrics.New()
	c.Record(metrics.Entry{Type: metrics.EventRotation})
	c.Record(metrics.Entry{Type: metrics.EventRotation})
	c.Record(metrics.Entry{Type: metrics.EventError})
	var buf bytes.Buffer
	r := metrics.NewReporter(c, &buf)
	r.PrintTable()
	out := buf.String()
	if !strings.Contains(out, "2") {
		t.Errorf("expected count 2 in output: %s", out)
	}
}

func TestPrintJSON_ValidJSON(t *testing.T) {
	c := metrics.New()
	c.Record(metrics.Entry{Type: metrics.EventSync, Provider: "gcp", Key: "api/key", Success: true})
	var buf bytes.Buffer
	r := metrics.NewReporter(c, &buf)
	if err := r.PrintJSON(); err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}
	var entries []metrics.Entry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != 1 || entries[0].Provider != "gcp" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestNewReporter_DefaultsToStdout(t *testing.T) {
	// Just ensure no panic when out is nil.
	c := metrics.New()
	r := metrics.NewReporter(c, nil)
	if r == nil {
		t.Error("expected non-nil reporter")
	}
}
