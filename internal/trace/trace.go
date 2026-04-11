// Package trace provides request/operation tracing for secret operations,
// allowing correlation of related actions across providers via a shared trace ID.
package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type contextKey struct{}

// Span represents a single traced operation.
type Span struct {
	TraceID   string            `json:"trace_id"`
	SpanID    string            `json:"span_id"`
	Operation string            `json:"operation"`
	Provider  string            `json:"provider"`
	Key       string            `json:"key"`
	StartedAt time.Time         `json:"started_at"`
	EndedAt   time.Time         `json:"ended_at,omitempty"`
	Error     string            `json:"error,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Tracer records spans for secret operations.
type Tracer struct {
	mu    sync.Mutex
	spans []Span
}

// New creates a new Tracer.
func New() *Tracer {
	return &Tracer{}
}

// Start begins a new span and returns a context carrying the trace ID.
func (t *Tracer) Start(ctx context.Context, operation, provider, key string) (context.Context, *Span) {
	traceID := traceIDFromContext(ctx)
	if traceID == "" {
		traceID = newID()
	}
	span := Span{
		TraceID:   traceID,
		SpanID:    newID(),
		Operation: operation,
		Provider:  provider,
		Key:       key,
		StartedAt: time.Now().UTC(),
		Meta:      make(map[string]string),
	}
	ctx = context.WithValue(ctx, contextKey{}, traceID)
	return ctx, &span
}

// Finish marks the span as complete and records it.
func (t *Tracer) Finish(span *Span, err error) {
	span.EndedAt = time.Now().UTC()
	if err != nil {
		span.Error = err.Error()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, *span)
}

// Spans returns a copy of all recorded spans.
func (t *Tracer) Spans() []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Span, len(t.spans))
	copy(out, t.spans)
	return out
}

// Reset clears all recorded spans.
func (t *Tracer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = nil
}

func traceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
