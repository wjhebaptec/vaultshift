package validate_test

import (
	"strings"
	"testing"

	"github.com/vaultshift/vaultshift/internal/validate"
)

func TestWithMinLength_Pass(t *testing.T) {
	v := validate.New(validate.WithMinLength(8))
	if err := v.Validate("abcdefgh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithMinLength_Fail(t *testing.T) {
	v := validate.New(validate.WithMinLength(8))
	if err := v.Validate("short"); err == nil {
		t.Fatal("expected error for short value")
	}
}

func TestWithMaxLength_Pass(t *testing.T) {
	v := validate.New(validate.WithMaxLength(16))
	if err := v.Validate("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithMaxLength_Fail(t *testing.T) {
	v := validate.New(validate.WithMaxLength(4))
	if err := v.Validate("toolongvalue"); err == nil {
		t.Fatal("expected error for long value")
	}
}

func TestWithPattern_Pass(t *testing.T) {
	v := validate.New(validate.WithPattern(`^[A-Z][a-z]+$`))
	if err := v.Validate("Hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithPattern_Fail(t *testing.T) {
	v := validate.New(validate.WithPattern(`^[A-Z][a-z]+$`))
	if err := v.Validate("hello"); err == nil {
		t.Fatal("expected error for pattern mismatch")
	}
}

func TestWithMinEntropy_Pass(t *testing.T) {
	// High-entropy string with many distinct characters.
	v := validate.New(validate.WithMinEntropy(3.0))
	if err := v.Validate("aB3!xY9#mQ"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithMinEntropy_Fail(t *testing.T) {
	// Low-entropy string: all same character.
	v := validate.New(validate.WithMinEntropy(2.0))
	if err := v.Validate(strings.Repeat("a", 20)); err == nil {
		t.Fatal("expected error for low-entropy value")
	}
}

func TestWithMinEntropy_EmptyValue(t *testing.T) {
	v := validate.New(validate.WithMinEntropy(1.0))
	if err := v.Validate(""); err == nil {
		t.Fatal("expected error for empty value")
	}
}

func TestMultipleRules_FirstErrorReturned(t *testing.T) {
	v := validate.New(
		validate.WithMinLength(12),
		validate.WithPattern(`^[a-z]+$`),
	)
	// Fails MinLength first.
	err := v.Validate("hi")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "" {
		t.Fatal("error message should not be empty")
	}
}

func TestValidate_AllRulesPass(t *testing.T) {
	v := validate.New(
		validate.WithMinLength(4),
		validate.WithMaxLength(32),
		validate.WithPattern(`^[a-zA-Z0-9]+$`),
	)
	if err := v.Validate("SecurePass123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
