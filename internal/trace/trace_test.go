package trace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/trace"
)

func TestStart_CreatesSpanWithIDs(t *testing.T) {
	tr := trace.New()
	ctx := context.Background()
	_, span := tr.Start(ctx, "rotate", "aws", "db/password")

	if span.TraceID == "" {
		t.Error("expected non-empty TraceID")
	}
	if span.SpanID == "" {
		t.Error("expected non-empty SpanID")
	}
	if span.Operation != "rotate" {
		t.Errorf("expected operation 'rotate', got %q", span.Operation)
	}
	if span.Provider != "aws" {
		t.Errorf("expected provider 'aws', got %q", span.Provider)
	}
	if span.Key != "db/password" {
		t.Errorf("expected key 'db/password', got %q", span.Key)
	}
	if span.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
}

func TestStart_PropagatesTraceID(t *testing.T) {
	tr := trace.New()
	ctx := context.Background()
	ctx1, span1 := tr.Start(ctx, "rotate", "aws", "key1")
	_, span2 := tr.Start(ctx1, "sync", "gcp", "key2")

	if span1.TraceID != span2.TraceID {
		t.Errorf("expected same TraceID, got %q and %q", span1.TraceID, span2.TraceID)
	}
	if span1.SpanID == span2.SpanID {
		t.Error("expected distinct SpanIDs")
	}
}

func TestFinish_SetsEndedAt(t *testing.T) {
	tr := trace.New()
	_, span := tr.Start(context.Background(), "get", "vault", "secret")
	tr.Finish(span, nil)

	if span.EndedAt.IsZero() {
		t.Error("expected EndedAt to be set after Finish")
	}
	if span.Error != "" {
		t.Errorf("expected no error, got %q", span.Error)
	}
}

func TestFinish_RecordsError(t *testing.T) {
	tr := trace.New()
	_, span := tr.Start(context.Background(), "put", "aws", "token")
	tr.Finish(span, errors.New("access denied"))

	if span.Error != "access denied" {
		t.Errorf("expected error 'access denied', got %q", span.Error)
	}
}

func TestSpans_ReturnsAllRecorded(t *testing.T) {
	tr := trace.New()
	for i := 0; i < 3; i++ {
		_, sp := tr.Start(context.Background(), "list", "gcp", "key")
		tr.Finish(sp, nil)
	}

	spans := tr.Spans()
	if len(spans) != 3 {
		t.Errorf("expected 3 spans, got %d", len(spans))
	}
}

func TestReset_ClearsSpans(t *testing.T) {
	tr := trace.New()
	_, sp := tr.Start(context.Background(), "delete", "vault", "old")
	tr.Finish(sp, nil)
	tr.Reset()

	if len(tr.Spans()) != 0 {
		t.Error("expected spans to be empty after Reset")
	}
}

func TestSpan_MetaIsInitialized(t *testing.T) {
	tr := trace.New()
	_, span := tr.Start(context.Background(), "rotate", "aws", "key")
	span.Meta["region"] = "us-east-1"

	if span.Meta["region"] != "us-east-1" {
		t.Error("expected meta to be writable")
	}
}
