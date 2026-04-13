// Package floor enforces a minimum length on secret values by padding
// short values up to a required minimum. This is useful when downstream
// systems require secrets to meet a minimum length constraint.
package floor

import (
	"errors"
	"strings"
)

// ErrInvalidMinLen is returned when the minimum length is not positive.
var ErrInvalidMinLen = errors.New("floor: minLen must be greater than zero")

// Option configures a Floor instance.
type Option func(*Floor)

// WithPadChar sets the character used to pad short values.
// Defaults to a space character.
func WithPadChar(ch rune) Option {
	return func(f *Floor) {
		f.padChar = ch
	}
}

// WithReject causes Apply to return an error instead of padding when the
// value is shorter than minLen.
func WithReject() Option {
	return func(f *Floor) {
		f.reject = true
	}
}

// Floor pads or rejects values that are shorter than a minimum length.
type Floor struct {
	minLen  int
	padChar rune
	reject  bool
}

// New creates a Floor with the given minimum length.
func New(minLen int, opts ...Option) (*Floor, error) {
	if minLen <= 0 {
		return nil, ErrInvalidMinLen
	}
	f := &Floor{
		minLen:  minLen,
		padChar: ' ',
	}
	for _, o := range opts {
		o(f)
	}
	return f, nil
}

// Apply ensures value meets the minimum length requirement.
// If the value is already long enough it is returned unchanged.
// If reject mode is enabled and the value is too short, an error is returned.
// Otherwise the value is right-padded to minLen with the pad character.
func (f *Floor) Apply(value string) (string, error) {
	if len(value) >= f.minLen {
		return value, nil
	}
	if f.reject {
		return "", errors.New("floor: value is shorter than minimum length")
	}
	padding := strings.Repeat(string(f.padChar), f.minLen-len(value))
	return value + padding, nil
}

// ApplyAll applies Apply to every value in the map, returning the first error
// encountered.
func (f *Floor) ApplyAll(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		padded, err := f.Apply(v)
		if err != nil {
			return nil, err
		}
		out[k] = padded
	}
	return out, nil
}
