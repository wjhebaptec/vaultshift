package export_test

import (
	"testing"

	"github.com/vaultshift/internal/export"
)

func TestWithPrefixFilter_Match(t *testing.T) {
	f := export.WithPrefixFilter("DB_")
	if !f("DB_PASSWORD") {
		t.Error("expected DB_PASSWORD to match prefix DB_")
	}
	if f("API_KEY") {
		t.Error("expected API_KEY to not match prefix DB_")
	}
}

func TestWithPrefixFilter_Empty(t *testing.T) {
	f := export.WithPrefixFilter("")
	if !f("ANYTHING") {
		t.Error("empty prefix should match all keys")
	}
}

func TestWithExcludeFilter_Excludes(t *testing.T) {
	f := export.WithExcludeFilter("SECRET", "PRIVATE")
	if f("DB_SECRET_KEY") {
		t.Error("expected DB_SECRET_KEY to be excluded")
	}
	if f("PRIVATE_TOKEN") {
		t.Error("expected PRIVATE_TOKEN to be excluded")
	}
	if !f("API_KEY") {
		t.Error("expected API_KEY to be included")
	}
}

func TestChainFilters_AllMustPass(t *testing.T) {
	prefix := export.WithPrefixFilter("DB_")
	exclude := export.WithExcludeFilter("OLD")
	chained := export.ChainFilters(prefix, exclude)

	if !chained("DB_PASSWORD") {
		t.Error("DB_PASSWORD should pass both filters")
	}
	if chained("DB_OLD_PASSWORD") {
		t.Error("DB_OLD_PASSWORD should be excluded by exclude filter")
	}
	if chained("API_KEY") {
		t.Error("API_KEY should fail prefix filter")
	}
}

func TestChainFilters_NoFilters(t *testing.T) {
	chained := export.ChainFilters()
	if !chained("ANY_KEY") {
		t.Error("chain with no filters should pass everything")
	}
}

func TestWithExcludeFilter_NoSubstrings(t *testing.T) {
	f := export.WithExcludeFilter()
	if !f("ANY_KEY") {
		t.Error("exclude filter with no substrings should pass everything")
	}
}
