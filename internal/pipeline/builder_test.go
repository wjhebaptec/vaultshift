package pipeline_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vaultshift/internal/pipeline"
)

func TestBuilder_WithValidation_Pass(t *testing.T) {
	p := pipeline.NewBuilder().
		WithValidation("non-empty", func(v string) bool { return v != "" }).
		Build()

	_, err := p.Execute(context.Background(), newPayload("k", "aws", "secret"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuilder_WithValidation_Fail(t *testing.T) {
	p := pipeline.NewBuilder().
		WithValidation("non-empty", func(v string) bool { return v != "" }).
		Build()

	_, err := p.Execute(context.Background(), newPayload("k", "aws", ""))
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "non-empty") {
		t.Errorf("error should mention validation name, got: %v", err)
	}
}

func TestBuilder_WithTransform_Applies(t *testing.T) {
	p := pipeline.NewBuilder().
		WithTransform("upper", func(v string) (string, error) { return strings.ToUpper(v), nil }).
		Build()

	pl := newPayload("k", "gcp", "hello")
	_, err := p.Execute(context.Background(), pl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl.Value != "HELLO" {
		t.Errorf("expected HELLO, got %q", pl.Value)
	}
}

func TestBuilder_WithTransform_Error(t *testing.T) {
	p := pipeline.NewBuilder().
		WithTransform("fail", func(_ string) (string, error) { return "", errors.New("transform error") }).
		Build()

	_, err := p.Execute(context.Background(), newPayload("k", "vault", "v"))
	if err == nil {
		t.Fatal("expected error from transform")
	}
}

func TestBuilder_WithMetaTag_SetsValue(t *testing.T) {
	p := pipeline.NewBuilder().
		WithMetaTag("env", "production").
		Build()

	pl := newPayload("k", "aws", "v")
	_, err := p.Execute(context.Background(), pl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl.Meta["env"] != "production" {
		t.Errorf("expected meta env=production, got %q", pl.Meta["env"])
	}
}

func TestBuilder_Chained(t *testing.T) {
	pl := newPayload("db/pass", "aws", "weak")
	p := pipeline.NewBuilder().
		WithValidation("non-empty", func(v string) bool { return len(v) > 0 }).
		WithTransform("prefix", func(v string) (string, error) { return "rotated-" + v, nil }).
		WithMetaTag("rotated", "true").
		Build()

	_, err := p.Execute(context.Background(), pl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(pl.Value, "rotated-") {
		t.Errorf("expected rotated- prefix, got %q", pl.Value)
	}
	if pl.Meta["rotated"] != "true" {
		t.Error("expected meta rotated=true")
	}
}
