package scrub_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/provider/mock"
	"github.com/vaultshift/internal/scrub"
)

func TestClean_NoRules_ReturnsOriginal(t *testing.T) {
	s := scrub.New()
	got := s.Clean(context.Background(), "super-secret")
	if got != "super-secret" {
		t.Fatalf("expected original value, got %q", got)
	}
}

func TestClean_PatternReplaced(t *testing.T) {
	s := scrub.New()
	if err := s.AddRule(`\d{4}`, "****"); err != nil {
		t.Fatal(err)
	}
	got := s.Clean(context.Background(), "pin:1234")
	if got != "pin:****" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestClean_MultipleRulesApplied(t *testing.T) {
	s := scrub.New()
	_ = s.AddRule(`secret`, "[REDACTED]")
	_ = s.AddRule(`\d+`, "#")
	got := s.Clean(context.Background(), "secret123")
	if got != "[REDACTED]#" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestAddRule_EmptyPattern_ReturnsError(t *testing.T) {
	s := scrub.New()
	if err := s.AddRule("", "x"); err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestAddRule_InvalidRegex_ReturnsError(t *testing.T) {
	s := scrub.New()
	if err := s.AddRule(`[invalid`, "x"); err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestWrap_NilProvider_ReturnsError(t *testing.T) {
	s := scrub.New()
	if _, err := scrub.Wrap(nil, s); err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestWrap_NilScrubber_ReturnsError(t *testing.T) {
	p := mock.New("test")
	if _, err := scrub.Wrap(p, nil); err == nil {
		t.Fatal("expected error for nil scrubber")
	}
}

func TestWrap_PutScrubsValue(t *testing.T) {
	ctx := context.Background()
	m := mock.New("test")
	s := scrub.New()
	_ = s.AddRule(`password=\S+`, "password=[SCRUBBED]")

	p, err := scrub.Wrap(m, s)
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Put(ctx, "creds", "user=admin password=hunter2"); err != nil {
		t.Fatal(err)
	}

	got, err := p.Get(ctx, "creds")
	if err != nil {
		t.Fatal(err)
	}
	if got != "user=admin password=[SCRUBBED]" {
		t.Fatalf("unexpected stored value: %q", got)
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	s := scrub.New()
	_ = s.AddRule(`foo`, "bar")
	r1 := s.Rules()
	r2 := s.Rules()
	if len(r1) != 1 || len(r2) != 1 {
		t.Fatal("expected one rule")
	}
}
