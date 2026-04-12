package trim_test

import (
	"strings"
	"testing"

	"github.com/vaultshift/internal/trim"
)

func TestNew_InvalidMaxLen_ReturnsError(t *testing.T) {
	_, err := trim.New(0)
	if err == nil {
		t.Fatal("expected error for maxLen=0")
	}
	_, err = trim.New(-5)
	if err == nil {
		t.Fatal("expected error for negative maxLen")
	}
}

func TestTrim_ShortValue_Unchanged(t *testing.T) {
	tr, _ := trim.New(10)
	got := tr.Trim("hello")
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestTrim_ExactLength_Unchanged(t *testing.T) {
	tr, _ := trim.New(5)
	got := tr.Trim("hello")
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestTrim_LongValue_Truncated(t *testing.T) {
	tr, _ := trim.New(4)
	got := tr.Trim("abcdefgh")
	if got != "abcd" {
		t.Fatalf("expected 'abcd', got %q", got)
	}
}

func TestTrim_WithSuffix_AppendsSuffix(t *testing.T) {
	tr, _ := trim.New(8, trim.WithSuffix("..."))
	got := tr.Trim("supersecretvalue")
	if len(got) != 8 {
		t.Fatalf("expected length 8, got %d", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected suffix '...', got %q", got)
	}
}

func TestTrim_WithSuffix_NoTruncation_NoSuffix(t *testing.T) {
	tr, _ := trim.New(20, trim.WithSuffix("..."))
	got := tr.Trim("short")
	if got != "short" {
		t.Fatalf("expected 'short', got %q", got)
	}
}

func TestTrim_SuffixLargerThanMax_ClampsToMax(t *testing.T) {
	tr, _ := trim.New(2, trim.WithSuffix("..."))
	got := tr.Trim("abcdefgh")
	if len(got) > 2 {
		t.Fatalf("expected at most 2 bytes, got %d", len(got))
	}
}

func TestTrimAll_AppliestoAllKeys(t *testing.T) {
	tr, _ := trim.New(3)
	secrets := map[string]string{
		"a": "hello",
		"b": "hi",
		"c": "toolongvalue",
	}
	out := tr.TrimAll(secrets)
	for k, v := range out {
		if len(v) > 3 {
			t.Errorf("key %q: expected len <= 3, got %d", k, len(v))
		}
	}
	if out["b"] != "hi" {
		t.Errorf("expected 'hi' unchanged, got %q", out["b"])
	}
}

func TestTruncated_ReportsCorrectly(t *testing.T) {
	tr, _ := trim.New(5)
	if tr.Truncated("hi") {
		t.Error("expected false for short value")
	}
	if !tr.Truncated("toolong") {
		t.Error("expected true for long value")
	}
}
