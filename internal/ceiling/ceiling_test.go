package ceiling_test

import (
	"errors"
	"testing"

	"github.com/vaultshift/internal/ceiling"
)

func TestNew_InvalidMaxLen_ReturnsError(t *testing.T) {
	_, err := ceiling.New(0)
	if err == nil {
		t.Fatal("expected error for maxLen=0")
	}
}

func TestApply_ShortValue_Unchanged(t *testing.T) {
	c, _ := ceiling.New(10)
	out, err := c.Apply("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected 'hello', got %q", out)
	}
}

func TestApply_ExactLength_Unchanged(t *testing.T) {
	c, _ := ceiling.New(5)
	out, err := c.Apply("hello")
	if err != nil || out != "hello" {
		t.Errorf("expected unchanged value, got %q, err=%v", out, err)
	}
}

func TestApply_LongValue_Truncated(t *testing.T) {
	c, _ := ceiling.New(4)
	out, err := c.Apply("toolong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "tool" {
		t.Errorf("expected 'tool', got %q", out)
	}
}

func TestApply_WithReject_ReturnsError(t *testing.T) {
	c, _ := ceiling.New(4, ceiling.WithReject())
	_, err := c.Apply("toolong")
	if !errors.Is(err, ceiling.ErrExceedsCeiling) {
		t.Errorf("expected ErrExceedsCeiling, got %v", err)
	}
}

func TestApply_WithReject_ShortValue_NoError(t *testing.T) {
	c, _ := ceiling.New(10, ceiling.WithReject())
	out, err := c.Apply("ok")
	if err != nil || out != "ok" {
		t.Errorf("expected 'ok', got %q, err=%v", out, err)
	}
}

func TestApplyAll_TruncatesAll(t *testing.T) {
	c, _ := ceiling.New(3)
	in := map[string]string{"a": "hello", "b": "hi", "c": "world"}
	out, err := c.ApplyAll(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["a"] != "hel" || out["b"] != "hi" || out["c"] != "wor" {
		t.Errorf("unexpected output: %v", out)
	}
}

func TestApplyAll_WithReject_StopsOnFirstError(t *testing.T) {
	c, _ := ceiling.New(3, ceiling.WithReject())
	in := map[string]string{"key": "toolong"}
	_, err := c.ApplyAll(in)
	if !errors.Is(err, ceiling.ErrExceedsCeiling) {
		t.Errorf("expected ErrExceedsCeiling, got %v", err)
	}
}
