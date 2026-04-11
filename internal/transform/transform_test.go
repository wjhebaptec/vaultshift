package transform_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/vaultshift/internal/transform"
)

func TestApply_SingleStep(t *testing.T) {
	tr := transform.New(transform.TrimSpace())
	got, err := tr.Apply("  hello  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestApply_ChainedSteps(t *testing.T) {
	tr := transform.New(transform.TrimSpace(), transform.ToUpper(), transform.AddPrefix("KEY_"))
	got, err := tr.Apply("  secret  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "KEY_SECRET" {
		t.Errorf("expected %q, got %q", "KEY_SECRET", got)
	}
}

func TestApply_StopsOnError(t *testing.T) {
	failStep := func(_ string) (string, error) {
		return "", errors.New("step failed")
	}
	called := false
	afterStep := func(v string) (string, error) {
		called = true
		return v, nil
	}
	tr := transform.New(failStep, afterStep)
	_, err := tr.Apply("value")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if called {
		t.Error("step after failure should not have been called")
	}
}

func TestApplyAll_TransformsMap(t *testing.T) {
	tr := transform.New(transform.ToLower())
	secrets := map[string]string{"A": "HELLO", "B": "WORLD"}
	got, err := tr.ApplyAll(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["A"] != "hello" || got["B"] != "world" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestApplyAll_ReturnsErrorOnFailure(t *testing.T) {
	failStep := func(_ string) (string, error) { return "", errors.New("oops") }
	tr := transform.New(failStep)
	_, err := tr.ApplyAll(map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBase64Encode_Decode_RoundTrip(t *testing.T) {
	original := "super-secret-value"
	encoder := transform.New(transform.Base64Encode())
	encoded, err := encoder.Apply(original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(encoded); err != nil {
		t.Fatalf("encoded value is not valid base64: %v", err)
	}
	decoder := transform.New(transform.Base64Decode())
	decoded, err := decoder.Apply(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if decoded != original {
		t.Errorf("expected %q, got %q", original, decoded)
	}
}

func TestBase64Decode_InvalidInput(t *testing.T) {
	tr := transform.New(transform.Base64Decode())
	_, err := tr.Apply("!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}
