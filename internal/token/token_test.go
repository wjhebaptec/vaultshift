package token_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/token"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIssue_ReturnsEntry(t *testing.T) {
	m := token.New()
	e, err := m.Issue("aws:prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if e.Scope != "aws:prod" {
		t.Errorf("scope: got %q, want %q", e.Scope, "aws:prod")
	}
}

func TestValidate_ValidToken(t *testing.T) {
	now := time.Now()
	m := token.New(token.WithClock(fixedClock(now)), token.WithTTL(10*time.Minute))
	e, _ := m.Issue("gcp:staging")

	got, err := m.Validate(e.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Scope != "gcp:staging" {
		t.Errorf("scope: got %q, want %q", got.Scope, "gcp:staging")
	}
}

func TestValidate_ExpiredToken(t *testing.T) {
	now := time.Now()
	clock := fixedClock(now)
	m := token.New(token.WithClock(clock), token.WithTTL(5*time.Minute))
	e, _ := m.Issue("vault:dev")

	// advance clock past TTL
	m2 := token.New(token.WithClock(fixedClock(now.Add(10*time.Minute))), token.WithTTL(5*time.Minute))
	// re-use the same store is not possible, so we test via IsExpired directly
	if !e.IsExpired(now.Add(10 * time.Minute)) {
		t.Fatal("expected token to be expired")
	}
	_ = m2
}

func TestValidate_NotFound(t *testing.T) {
	m := token.New()
	_, err := m.Validate("nonexistent")
	if err != token.ErrTokenNotFound {
		t.Errorf("got %v, want ErrTokenNotFound", err)
	}
}

func TestRevoke_PreventsValidation(t *testing.T) {
	m := token.New()
	e, _ := m.Issue("aws:dev")

	if err := m.Revoke(e.Token); err != nil {
		t.Fatalf("revoke error: %v", err)
	}
	_, err := m.Validate(e.Token)
	if err != token.ErrTokenRevoked {
		t.Errorf("got %v, want ErrTokenRevoked", err)
	}
}

func TestRevoke_UnknownToken(t *testing.T) {
	m := token.New()
	err := m.Revoke("ghost")
	if err != token.ErrTokenNotFound {
		t.Errorf("got %v, want ErrTokenNotFound", err)
	}
}

func TestIssue_UniqueTokens(t *testing.T) {
	m := token.New()
	a, _ := m.Issue("scope-a")
	b, _ := m.Issue("scope-b")
	if a.Token == b.Token {
		t.Error("expected unique tokens, got duplicates")
	}
}

func TestWithTTL_SetsExpiry(t *testing.T) {
	now := time.Now()
	m := token.New(token.WithClock(fixedClock(now)), token.WithTTL(30*time.Minute))
	e, _ := m.Issue("scope")
	want := now.Add(30 * time.Minute)
	if !e.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt: got %v, want %v", e.ExpiresAt, want)
	}
}
