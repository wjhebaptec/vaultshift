// Package sanitize provides utilities for cleaning and normalizing secret
// values before they are stored or transmitted across providers.
package sanitize

import (
	"strings"
	"unicode"
)

// Option configures a Sanitizer.
type Option func(*Sanitizer)

// Sanitizer applies a chain of cleaning steps to secret values.
type Sanitizer struct {
	steps []func(string) string
}

// WithTrimSpace adds a step that removes leading and trailing whitespace.
func WithTrimSpace() Option {
	return func(s *Sanitizer) {
		s.steps = append(s.steps, strings.TrimSpace)
	}
}

// WithStripControl adds a step that removes non-printable control characters.
func WithStripControl() Option {
	return func(s *Sanitizer) {
		s.steps = append(s.steps, func(v string) string {
			return strings.Map(func(r rune) rune {
				if unicode.IsControl(r) {
					return -1
				}
				return r
			}, v)
		})
	}
}

// WithCollapseSpaces adds a step that replaces consecutive whitespace with a
// single space.
func WithCollapseSpaces() Option {
	return func(s *Sanitizer) {
		s.steps = append(s.steps, func(v string) string {
			fields := strings.Fields(v)
			return strings.Join(fields, " ")
		})
	}
}

// WithNullToEmpty adds a step that converts the literal string "null" (case-
// insensitive) to an empty string.
func WithNullToEmpty() Option {
	return func(s *Sanitizer) {
		s.steps = append(s.steps, func(v string) string {
			if strings.EqualFold(v, "null") {
				return ""
			}
			return v
		})
	}
}

// New creates a Sanitizer with the given options.
func New(opts ...Option) *Sanitizer {
	s := &Sanitizer{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Clean applies all configured steps to value and returns the result.
func (s *Sanitizer) Clean(value string) string {
	for _, step := range s.steps {
		value = step(value)
	}
	return value
}

// CleanMap applies Clean to every value in the provided map and returns a new
// map with the sanitized values.
func (s *Sanitizer) CleanMap(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = s.Clean(v)
	}
	return out
}
