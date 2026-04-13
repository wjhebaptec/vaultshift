package clamp_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/vaultshift/internal/clamp"
)

func TestNew_InvalidMin_ReturnsError(t *testing.T) {
	_, err := clamp.New(-1, 10)
	if err == nil {
		t.Fatal("expected error for negative min")
	}
}

func TestNew_MaxLessThanMin_ReturnsError(t *testing.T) {
	_, err := clamp.New(10, 5)
	if err == nil {
		t.Fatal("expected error when max < min")
	}
}

func TestApply_WithinRange_Unchanged(t *testing.T) {
	c, _ := clamp.New(3, 10)
	out, err := c.Apply("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Fatalf("expected 'hello', got %q", out)
	}
}

func TestApply_TooShort_PadsWithSpaces(t *testing.T) {
	c, _ := clamp.New(8, 20)
	out, err := c.Apply("hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 8 {
		t.Fatalf("expected length 8, got %d", len(out))
	}
	if !strings.HasPrefix(out, "hi") {
		t.Fatalf("expected prefix 'hi', got %q", out)
	}
}

func TestApply_TooLong_Truncated(t *testing.T) {
	c, _ := clamp.New(1, 5)
	out, err := c.Apply("toolongvalue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "toolo" {
		t.Fatalf("expected 'toolo', got %q", out)
	}
}

func TestApply_WithReject_TooLong_ReturnsError(t *testing.T) {
	c, _ := clamp.New(1, 5, clamp.WithReject())
	_, err := c.Apply("toolongvalue")
	if !errors.Is(err, clamp.ErrOutOfRange) {
		t.Fatalf("expected ErrOutOfRange, got %v", err)
	}
}

func TestApply_WithReject_TooShort_ReturnsError(t *testing.T) {
	c, _ := clamp.New(5, 20, clamp.WithReject())
	_, err := c.Apply("ab")
	if !errors.Is(err, clamp.ErrOutOfRange) {
		t.Fatalf("expected ErrOutOfRange, got %v", err)
	}
}

func TestApplyAll_AllWithinRange(t *testing.T) {
	c, _ := clamp.New(1, 10)
	m := map[string]string{"a": "foo", "b": "bar"}
	out, err := c.ApplyAll(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestApplyAll_WithReject_StopsOnFirstError(t *testing.T) {
	c, _ := clamp.New(1, 3, clamp.WithReject())
	m := map[string]string{"key": "toolong"}
	_, err := c.ApplyAll(m)
	if err == nil {
		t.Fatal("expected error from ApplyAll")
	}
}
