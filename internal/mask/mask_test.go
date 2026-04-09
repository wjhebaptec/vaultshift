package mask

import (
	"strings"
	"testing"
)

func TestMask_EmptyString(t *testing.T) {
	m := New()
	if got := m.Mask(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestMask_ShortValue_FullyMasked(t *testing.T) {
	m := New() // visibleSuffix=4
	// value shorter than or equal to suffix should be fully masked
	got := m.Mask("abc")
	if got != "***" {
		t.Fatalf("expected ***, got %q", got)
	}
}

func TestMask_LongValue_SuffixVisible(t *testing.T) {
	m := New()
	got := m.Mask("supersecret1234")
	if !strings.HasSuffix(got, "1234") {
		t.Fatalf("expected suffix '1234' to be visible, got %q", got)
	}
	if strings.Contains(got[:len(got)-4], "s") {
		t.Fatalf("expected hidden portion to be masked, got %q", got)
	}
}

func TestMask_LengthPreserved(t *testing.T) {
	m := New()
	value := "my-secret-value"
	got := m.Mask(value)
	if len(got) != len(value) {
		t.Fatalf("masked length %d != original length %d", len(got), len(value))
	}
}

func TestMask_WithFullMask(t *testing.T) {
	m := New(WithFullMask())
	got := m.Mask("visible1.Contains(got, "1234") {
		t.Fatalf("expected no visible characters, got %q", got)
	}
	for _, r := range got {
		if r != DefaultMaskChar {
			t.Fatalf("unexpected character %q in fully-masked output", r)
		}
	}
}

func TestMask_CustomMaskChar(t *testing.T) {
	m := New(WithMaskChar('#'), WithFullMask())
	got := m.Mask("hello")
	if got != "#####" {
		t.Fatalf("expected #####, got %q", got)
	}
}

func TestMask_CustomVisibleSuffix(t *testing.T) {
	m := New(WithVisibleSuffix(2))
	got := m.Mask("password99")
	if !strings.HasSuffix(got, "99") {
		t.Fatalf("expected suffix '99', got %q", got)
	}
}

func TestMaskAll_MasksAllValues(t *testing.T) {
	m := New(WithFullMask())
	secrets := map[string]string{
		"db_pass": "hunter2",
		"api_key": "abc123xyz",
	}
	masked := m.MaskAll(secrets)
	for k, v := range masked {
		for _, r := range v {
			if r != DefaultMaskChar {
				t.Fatalf("key %q: unexpected unmasked character in %q", k, v)
			}
		}
		if len(v) != len(secrets[k]) {
			t.Fatalf("key %q: length mismatch", k)
		}
	}
}

func TestMaskAll_DoesNotMutateOriginal(t *testing.T) {
	m := New()
	orig := map[string]string{"key": "plaintext"}
	_ = m.MaskAll(orig)
	if orig["key"] != "plaintext" {
		t.Fatal("original map was mutated")
	}
}
