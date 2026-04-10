package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/pipeline"
)

func newPayload(key, provider, value string) *pipeline.Payload {
	return &pipeline.Payload{Key: key, Provider: provider, Value: value, Meta: map[string]string{}}
}

func TestExecute_AllStepsRun(t *testing.T) {
	order := []string{}
	p := pipeline.New().
		Add(pipeline.Step{Name: "a", Run: func(_ context.Context, _ *pipeline.Payload) error { order = append(order, "a"); return nil }}).
		Add(pipeline.Step{Name: "b", Run: func(_ context.Context, _ *pipeline.Payload) error { order = append(order, "b"); return nil }})

	results, err := p.Execute(context.Background(), newPayload("k", "aws", "v"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if order[0] != "a" || order[1] != "b" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestExecute_StopsOnError(t *testing.T) {
	ran := false
	p := pipeline.New().
		Add(pipeline.Step{Name: "fail", Run: func(_ context.Context, _ *pipeline.Payload) error { return errors.New("boom") }}).
		Add(pipeline.Step{Name: "never", Run: func(_ context.Context, _ *pipeline.Payload) error { ran = true; return nil }})

	results, err := p.Execute(context.Background(), newPayload("k", "gcp", "v"))
	if err == nil {
		t.Fatal("expected error")
	}
	if ran {
		t.Error("second step should not have run")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestExecute_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := pipeline.New().
		Add(pipeline.Step{Name: "cancel", Run: func(_ context.Context, _ *pipeline.Payload) error { cancel(); return nil }}).
		Add(pipeline.Step{Name: "after", Run: func(_ context.Context, _ *pipeline.Payload) error { return nil }})

	_, err := p.Execute(ctx, newPayload("k", "vault", "v"))
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestExecute_PayloadMutated(t *testing.T) {
	p := pipeline.New().
		Add(pipeline.Step{Name: "mutate", Run: func(_ context.Context, pl *pipeline.Payload) error { pl.Value = "rotated"; return nil }})

	payload := newPayload("mykey", "aws", "original")
	_, err := p.Execute(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Value != "rotated" {
		t.Errorf("expected value 'rotated', got %q", payload.Value)
	}
}

func TestExecute_EmptyPipeline(t *testing.T) {
	p := pipeline.New()
	results, err := p.Execute(context.Background(), newPayload("k", "aws", "v"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAdd_AutoName(t *testing.T) {
	p := pipeline.New().
		Add(pipeline.Step{Run: func(_ context.Context, _ *pipeline.Payload) error { return nil }})
	if p.Len() != 1 {
		t.Errorf("expected 1 step, got %d", p.Len())
	}
}
