package policy_test

import (
	"testing"
	"time"

	"github.com/vaultshift/vaultshift/internal/policy"
)

func TestNew_DefaultPolicy(t *testing.T) {
	p := policy.New("default")
	if p.Name != "default" {
		t.Fatalf("expected name \"default\", got %q", p.Name)
	}
	if p.MaxAge != 0 {
		t.Fatalf("expected zero MaxAge, got %v", p.MaxAge)
	}
	if p.AllowedPattern != nil {
		t.Fatal("expected nil AllowedPattern")
	}
}

func TestValidate_PatternMatch(t *testing.T) {
	p := policy.New("prod", policy.WithAllowedPattern(`^prod/`))

	if err := p.Validate("prod/db_pass", time.Time{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := p.Validate("dev/db_pass", time.Time{}); err == nil {
		t.Fatal("expected error for non-matching key, got nil")
	}
}

func TestValidate_MaxAge_Fresh(t *testing.T) {
	p := policy.New("ttl", policy.WithMaxAge(24*time.Hour))

	lastRotated := time.Now().Add(-1 * time.Hour)
	if err := p.Validate("my/secret", lastRotated); err != nil {
		t.Fatalf("expected no error for fresh secret, got: %v", err)
	}
}

func TestValidate_MaxAge_Expired(t *testing.T) {
	p := policy.New("ttl", policy.WithMaxAge(24*time.Hour))

	lastRotated := time.Now().Add(-48 * time.Hour)
	if err := p.Validate("my/secret", lastRotated); err == nil {
		t.Fatal("expected error for expired secret, got nil")
	}
}

func TestValidate_ZeroLastRotated_SkipsAgeCheck(t *testing.T) {
	p := policy.New("ttl", policy.WithMaxAge(1*time.Minute))

	// zero time means unknown rotation time — should not fail age check
	if err := p.Validate("any/key", time.Time{}); err != nil {
		t.Fatalf("expected no error for zero lastRotated, got: %v", err)
	}
}

func TestValidate_RequiredTargets_StoredOnPolicy(t *testing.T) {
	p := policy.New("multi", policy.WithRequiredTargets("aws", "gcp"))

	if len(p.RequiredTargets) != 2 {
		t.Fatalf("expected 2 required targets, got %d", len(p.RequiredTargets))
	}
	if p.RequiredTargets[0] != "aws" || p.RequiredTargets[1] != "gcp" {
		t.Fatalf("unexpected targets: %v", p.RequiredTargets)
	}
}

func TestValidate_CombinedOptions(t *testing.T) {
	p := policy.New("strict",
		policy.WithAllowedPattern(`^prod/`),
		policy.WithMaxAge(72*time.Hour),
		policy.WithRequiredTargets("aws"),
	)

	// valid key, fresh rotation
	if err := p.Validate("prod/api_key", time.Now().Add(-1*time.Hour)); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// invalid key pattern
	if err := p.Validate("staging/api_key", time.Now().Add(-1*time.Hour)); err == nil {
		t.Fatal("expected pattern error")
	}
}
