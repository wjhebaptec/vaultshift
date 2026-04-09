// Package mask provides utilities for redacting and masking sensitive
// secret values before they are logged, printed, or transmitted.
package mask

import (
	"strings"
	"unicode/utf8"
)

const (
	// DefaultMaskChar is the character used to replace secret characters.
	DefaultMaskChar = '*'
	// DefaultVisibleSuffix is the number of trailing characters to reveal.
	DefaultVisibleSuffix = 4
)

// Masker redacts secret values according to a configured policy.
type Masker struct {
	maskChar      rune
	visibleSuffix int
	fullMask      bool
}

// Option configures a Masker.
type Option func(*Masker)

// WithMaskChar sets the character used for masking.
func WithMaskChar(r rune) Option {
	return func(m *Masker) { m.maskChar = r }
}

// WithVisibleSuffix sets how many trailing characters remain visible.
func WithVisibleSuffix(n int) Option {
	return func(m *Masker) { m.visibleSuffix = n }
}

// WithFullMask forces every character to be masked regardless of length.
func WithFullMask() Option {
	return func(m *Masker) { m.fullMask = true }
}

// New creates a Masker with the supplied options.
func New(opts ...Option) *Masker {
	m := &Masker{
		maskChar:      DefaultMaskChar,
		visibleSuffix: DefaultVisibleSuffix,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Mask redacts value, optionally preserving a short suffix for identification.
func (m *Masker) Mask(value string) string {
	if value == "" {
		return ""
	}
	l := utf8.RuneCountInString(value)
	if m.fullMask || l <= m.visibleSuffix {
		return strings.Repeat(string(m.maskChar), l)
	}
	hidden := l - m.visibleSuffix
	suffix := string([]rune(value)[hidden:])
	return strings.Repeat(string(m.maskChar), hidden) + suffix
}

// MaskAll applies Mask to every value in the provided map, returning a new map.
func (m *Masker) MaskAll(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = m.Mask(v)
	}
	return out
}
