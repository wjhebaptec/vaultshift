package floor_test

import (
	"testing"

	"github.com/vaultshift/vaultshift/internal/floor"
)

func TestNew_InvalidMinLen_ReturnsError(t *testing.T) {
	_, err := floor.New(0)
	if err == nil {
		t.Fatal("expected error for zero minLen")
	}
}

func TestApply_ShortValue_PaddedWithSpaces(t *testing.T) {
	f, _ := floor.New(8)
	got, err := f.Apply("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 8 {
		t.Fatalf("expected length 8, got %d", len(got))
	}
	if got[:3] != "abc" {
		t.Fatalf("expected prefix 'abc', got %q", got[:3])
	}
}

func TestApply_ExactLength_Unchanged(t *testing.T) {
	f, _ := floor.New(5)
	got, err := f.Apply("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestApply_LongValue_Unchanged(t *testing.T) {
	f, _ := floor.New(3)
	got, err := f.Apply("toolong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "toolong" {
		t.Fatalf("expected 'toolong', got %q", got)
	}
}

func TestApply_WithPadChar_UsesCustomChar(t *testing.T) {
	f, _ := floor.New(6, floor.WithPadChar('*'))
	got, err := f.Apply("hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hi****" {
		t.Fatalf("expected 'hi****', got %q", got)
	}
}

func TestApply_WithReject_ReturnsError(t *testing.T) {
	f, _ := floor.New(10, floor.WithReject())
	_, err := f.Apply("short")
	if err == nil {
		t.Fatal("expected error when value is too short in reject mode")
	}
}

func TestApply_WithReject_LongEnough_NoError(t *testing.T) {
	f, _ := floor.New(4, floor.WithReject())
	got, err := f.Apply("pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "pass" {
		t.Fatalf("expected 'pass', got %q", got)
	}
}

func TestApplyAll_PadsAllValues(t *testing.T) {
	f, _ := floor.New(5)
	secrets := map[string]string{"a": "hi", "b": "hello", "c": "x"}
	out, err := f.ApplyAll(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for k, v := range out {
		if len(v) < 5 {
			t.Errorf("key %q: expected length >= 5, got %d", k, len(v))
		}
	}
}

func TestApplyAll_WithReject_StopsOnFirstError(t *testing.T) {
	f, _ := floor.New(10, floor.WithReject())
	secrets := map[string]string{"a": "short"}
	_, err := f.ApplyAll(secrets)
	if err == nil {
		t.Fatal("expected error from ApplyAll in reject mode")
	}
}
