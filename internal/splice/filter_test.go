package splice_test

import (
	"testing"

	"github.com/vaultshift/internal/splice"
)

func TestWithSourcePrefix_Matches(t *testing.T) {
	f := splice.WithSourcePrefix("prod/")
	if !f("src", "prod/db") {
		t.Error("expected match")
	}
	if f("src", "dev/db") {
		t.Error("expected no match")
	}
}

func TestWithSourceSuffix_Matches(t *testing.T) {
	f := splice.WithSourceSuffix("_key")
	if !f("src", "api_key") {
		t.Error("expected match")
	}
	if f("src", "api_secret") {
		t.Error("expected no match")
	}
}

func TestWithProviderName_Filters(t *testing.T) {
	f := splice.WithProviderName("aws")
	if !f("aws", "key") {
		t.Error("expected match for aws")
	}
	if f("gcp", "key") {
		t.Error("expected no match for gcp")
	}
}

func TestChainFilters_BothMustPass(t *testing.T) {
	f := splice.ChainFilters(
		splice.WithSourcePrefix("prod/"),
		splice.WithProviderName("aws"),
	)
	if !f("aws", "prod/key") {
		t.Error("expected both filters to pass")
	}
	if f("gcp", "prod/key") {
		t.Error("expected provider filter to block")
	}
	if f("aws", "dev/key") {
		t.Error("expected prefix filter to block")
	}
}

func TestChainFilters_NoFilters_AllPass(t *testing.T) {
	f := splice.ChainFilters()
	if !f("any", "any/key") {
		t.Error("empty chain should always pass")
	}
}

func TestAllowed_NilFilter_ReturnsTrue(t *testing.T) {
	if !splice.Allowed(nil, "src", "key") {
		t.Error("nil filter should allow everything")
	}
}
