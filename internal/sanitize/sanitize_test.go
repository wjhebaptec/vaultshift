package sanitize_test

import (
	"testing"

	"github.com/vaultshift/internal/sanitize"
)

func TestClean_TrimSpace(t *testing.T) {
	s := sanitize.New(sanitize.WithTrimSpace())
	got := s.Clean("  secret  ")
	if got != "secret" {
		t.Fatalf("expected %q, got %q", "secret", got)
	}
}

func TestClean_StripControl(t *testing.T) {
	s := sanitize.New(sanitize.WithStripControl())
	got := s.Clean("sec\x00ret\x1b")
	if got != "secret" {
		t.Fatalf("expected %q, got %q", "secret", got)
	}
}

func TestClean_CollapseSpaces(t *testing.T) {
	s := sanitize.New(sanitize.WithCollapseSpaces())
	got := s.Clean("my   secret   value")
	if got != "my secret value" {
		t.Fatalf("expected %q, got %q", "my secret value", got)
	}
}

func TestClean_NullToEmpty(t *testing.T) {
	s := sanitize.New(sanitize.WithNullToEmpty())
	for _, input := range []string{"null", "NULL", "Null"} {
		got := s.Clean(input)
		if got != "" {
			t.Fatalf("input %q: expected empty string, got %q", input, got)
		}
	}
}

func TestClean_NullToEmpty_NonNull(t *testing.T) {
	s := sanitize.New(sanitize.WithNullToEmpty())
	got := s.Clean("notnull")
	if got != "notnull" {
		t.Fatalf("expected %q, got %q", "notnull", got)
	}
}

func TestClean_ChainedSteps(t *testing.T) {
	s := sanitize.New(
		sanitize.WithTrimSpace(),
		sanitize.WithStripControl(),
		sanitize.WithCollapseSpaces(),
	)
	got := s.Clean("  hello\x00   world  ")
	if got != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", got)
	}
}

func TestClean_NoSteps_ReturnsOriginal(t *testing.T) {
	s := sanitize.New()
	input := "  raw value\x00  "
	got := s.Clean(input)
	if got != input {
		t.Fatalf("expected unchanged value %q, got %q", input, got)
	}
}

func TestCleanMap_SanitizesAllValues(t *testing.T) {
	s := sanitize.New(sanitize.WithTrimSpace())
	secrets := map[string]string{
		"key1": "  val1  ",
		"key2": " val2 ",
	}
	result := s.CleanMap(secrets)
	for k, want := range map[string]string{"key1": "val1", "key2": "val2"} {
		if result[k] != want {
			t.Fatalf("key %q: expected %q, got %q", k, want, result[k])
		}
	}
}

func TestCleanMap_DoesNotMutateOriginal(t *testing.T) {
	s := sanitize.New(sanitize.WithTrimSpace())
	orig := map[string]string{"k": "  v  "}
	_ = s.CleanMap(orig)
	if orig["k"] != "  v  " {
		t.Fatal("original map was mutated")
	}
}
