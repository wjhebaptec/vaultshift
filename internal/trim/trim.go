// Package trim provides utilities for truncating secret values to a
// maximum byte length, optionally appending a suffix to indicate truncation.
package trim

import "errors"

// ErrMaxLenZero is returned when a zero or negative max length is configured.
var ErrMaxLenZero = errors.New("trim: maxLen must be greater than zero")

// Option configures a Trimmer.
type Option func(*Trimmer)

// WithSuffix sets a suffix appended to values that were truncated.
// The suffix itself counts toward maxLen, so effective content is
// maxLen - len(suffix) bytes.
func WithSuffix(s string) Option {
	return func(t *Trimmer) { t.suffix = s }
}

// Trimmer truncates string values that exceed a maximum length.
type Trimmer struct {
	maxLen int
	suffix string
}

// New creates a Trimmer that truncates values longer than maxLen bytes.
func New(maxLen int, opts ...Option) (*Trimmer, error) {
	if maxLen <= 0 {
		return nil, ErrMaxLenZero
	}
	t := &Trimmer{maxLen: maxLen}
	for _, o := range opts {
		o(t)
	}
	return t, nil
}

// Trim truncates value if it exceeds the configured maxLen.
// If a suffix is set and truncation occurs, the suffix replaces the
// trailing bytes so the total length equals maxLen.
func (t *Trimmer) Trim(value string) string {
	if len(value) <= t.maxLen {
		return value
	}
	if t.suffix == "" {
		return value[:t.maxLen]
	}
	cutAt := t.maxLen - len(t.suffix)
	if cutAt <= 0 {
		return t.suffix[:t.maxLen]
	}
	return value[:cutAt] + t.suffix
}

// TrimAll applies Trim to every value in the provided map, returning a new map.
func (t *Trimmer) TrimAll(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = t.Trim(v)
	}
	return out
}

// Truncated reports whether value would be truncated by this Trimmer.
func (t *Trimmer) Truncated(value string) bool {
	return len(value) > t.maxLen
}
