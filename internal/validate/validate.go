// Package validate provides secret value validation utilities for vaultshift.
// It supports length, entropy, and regex-based constraints.
package validate

import (
	"fmt"
	"math"
	"regexp"
	"unicode/utf8"
)

// Rule is a single validation constraint applied to a secret value.
type Rule func(value string) error

// Validator applies a set of rules to secret values.
type Validator struct {
	rules []Rule
}

// New creates a Validator with the provided rules.
func New(rules ...Rule) *Validator {
	return &Validator{rules: rules}
}

// Validate runs all rules against value and returns the first error encountered.
func (v *Validator) Validate(value string) error {
	for _, r := range v.rules {
		if err := r(value); err != nil {
			return err
		}
	}
	return nil
}

// WithMinLength returns a Rule that rejects values shorter than n characters.
func WithMinLength(n int) Rule {
	return func(value string) error {
		if utf8.RuneCountInString(value) < n {
			return fmt.Errorf("value length %d is below minimum %d", utf8.RuneCountInString(value), n)
		}
		return nil
	}
}

// WithMaxLength returns a Rule that rejects values longer than n characters.
func WithMaxLength(n int) Rule {
	return func(value string) error {
		if utf8.RuneCountInString(value) > n {
			return fmt.Errorf("value length %d exceeds maximum %d", utf8.RuneCountInString(value), n)
		}
		return nil
	}
}

// WithPattern returns a Rule that rejects values not matching the given regex.
func WithPattern(pattern string) Rule {
	re := regexp.MustCompile(pattern)
	return func(value string) error {
		if !re.MatchString(value) {
			return fmt.Errorf("value does not match required pattern %q", pattern)
		}
		return nil
	}
}

// WithMinEntropy returns a Rule that rejects values whose Shannon entropy is
// below the given threshold (bits per character).
func WithMinEntropy(threshold float64) Rule {
	return func(value string) error {
		if value == "" {
			return fmt.Errorf("entropy check failed: empty value")
		}
		freq := make(map[rune]float64)
		for _, ch := range value {
			freq[ch]++
		}
		n := float64(utf8.RuneCountInString(value))
		var entropy float64
		for _, count := range freq {
			p := count / n
			entropy -= p * math.Log2(p)
		}
		if entropy < threshold {
			return fmt.Errorf("entropy %.2f is below minimum %.2f", entropy, threshold)
		}
		return nil
	}
}
