package sieve

import (
	"testing"
)

func TestNew_NilRule_ReturnsError(t *testing.T) {
	_, err := New(MatchPrefix("x"), nil)
	if err == nil {
		t.Fatal("expected error for nil rule")
	}
}

func TestAllow_NoRules_AlwaysTrue(t *testing.T) {
	s, _ := New()
	if !s.Allow("any-key") {
		t.Fatal("empty sieve should allow everything")
	}
}

func TestAllow_PrefixRule_Matches(t *testing.T) {
	s, _ := New(MatchPrefix("prod/"))
	if !s.Allow("prod/db-pass") {
		t.Fatal("expected prod/db-pass to pass")
	}
	if s.Allow("staging/db-pass") {
		t.Fatal("expected staging/db-pass to be blocked")
	}
}

func TestAllow_SuffixRule_Matches(t *testing.T) {
	s, _ := New(MatchSuffix("_KEY"))
	if !s.Allow("API_KEY") {
		t.Fatal("expected API_KEY to pass")
	}
	if s.Allow("API_SECRET") {
		t.Fatal("expected API_SECRET to be blocked")
	}
}

func TestAllow_RegexRule_Matches(t *testing.T) {
	rule, err := MatchRegex(`^secret/[a-z]+$`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, _ := New(rule)
	if !s.Allow("secret/alpha") {
		t.Fatal("expected match")
	}
	if s.Allow("secret/Alpha") {
		t.Fatal("expected no match for uppercase")
	}
}

func TestMatchRegex_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := MatchRegex(`[invalid`)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestDeny_InvertsRule(t *testing.T) {
	s, _ := New(Deny(MatchPrefix("internal/")))
	if s.Allow("internal/secret") {
		t.Fatal("expected internal/secret to be blocked")
	}
	if !s.Allow("public/secret") {
		t.Fatal("expected public/secret to pass")
	}
}

func TestFilter_ReturnsMatchingKeys(t *testing.T) {
	s, _ := New(MatchPrefix("prod/"))
	input := []string{"prod/a", "staging/b", "prod/c", "dev/d"}
	out := s.Filter(input)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if out[0] != "prod/a" || out[1] != "prod/c" {
		t.Fatalf("unexpected results: %v", out)
	}
}

func TestAllow_MultipleRules_AllMustPass(t *testing.T) {
	s, _ := New(MatchPrefix("prod/"), MatchSuffix("_KEY"))
	if !s.Allow("prod/API_KEY") {
		t.Fatal("expected prod/API_KEY to pass both rules")
	}
	if s.Allow("prod/API_SECRET") {
		t.Fatal("expected prod/API_SECRET to fail suffix rule")
	}
	if s.Allow("staging/API_KEY") {
		t.Fatal("expected staging/API_KEY to fail prefix rule")
	}
}
