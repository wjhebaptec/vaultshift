package absorb_test

import (
	"testing"

	"github.com/vaultshift/internal/absorb"
)

func TestWithSourcePrefix_Matches(t *testing.T) {
	f := absorb.WithSourcePrefix("prod/")
	if !f("prod/db") {
		t.Error("expected prod/db to match")
	}
	if f("staging/db") {
		t.Error("expected staging/db to not match")
	}
}

func TestWithSourceSuffix_Matches(t *testing.T) {
	f := absorb.WithSourceSuffix("_key")
	if !f("api_key") {
		t.Error("expected api_key to match")
	}
	if f("api_secret") {
		t.Error("expected api_secret to not match")
	}
}

func TestWithExclude_Excludes(t *testing.T) {
	f := absorb.WithExclude("tmp", "test")
	if f("tmp/token") {
		t.Error("expected tmp/token to be excluded")
	}
	if f("test_key") {
		t.Error("expected test_key to be excluded")
	}
	if !f("prod/token") {
		t.Error("expected prod/token to pass")
	}
}

func TestChainFilters_BothMustPass(t *testing.T) {
	f := absorb.ChainFilters(
		absorb.WithSourcePrefix("prod/"),
		absorb.WithSourceSuffix("_key"),
	)
	if !f("prod/api_key") {
		t.Error("expected prod/api_key to pass chain")
	}
	if f("prod/api_secret") {
		t.Error("expected prod/api_secret to fail chain")
	}
	if f("staging/api_key") {
		t.Error("expected staging/api_key to fail chain")
	}
}

func TestChainFilters_NoFilters_AllPass(t *testing.T) {
	f := absorb.ChainFilters()
	if !f("anything") {
		t.Error("expected empty chain to pass all keys")
	}
}

func TestAllowed_NilFilter_AlwaysTrue(t *testing.T) {
	if !absorb.Allowed(nil, "any-key") {
		t.Error("expected nil filter to allow all keys")
	}
}

func TestAllowed_FilterApplied(t *testing.T) {
	f := absorb.WithSourcePrefix("x/")
	if absorb.Allowed(f, "y/key") {
		t.Error("expected y/key to be blocked")
	}
	if !absorb.Allowed(f, "x/key") {
		t.Error("expected x/key to be allowed")
	}
}
