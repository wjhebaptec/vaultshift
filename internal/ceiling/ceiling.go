// Package ceiling enforces an upper bound on secret value length
// and optionally rejects values that exceed the configured maximum.
package ceiling

import (
	"errors"
	"fmt"
)

// ErrExceedsCeiling is returned when a value exceeds the configured ceiling.
var ErrExceedsCeiling = errors.New("value exceeds ceiling")

// Option configures a Ceiling.
type Option func(*Ceiling)

// WithReject causes the Ceiling to return an error instead of truncating.
func WithReject() Option {
	return func(c *Ceiling) {
		c.reject = true
	}
}

// Ceiling enforces a maximum byte length on secret values.
type Ceiling struct {
	maxLen int
	reject bool
}

// New creates a Ceiling with the given maximum length.
// Returns an error if maxLen is less than 1.
func New(maxLen int, opts ...Option) (*Ceiling, error) {
	if maxLen < 1 {
		return nil, fmt.Errorf("ceiling: maxLen must be at least 1, got %d", maxLen)
	}
	c := &Ceiling{maxLen: maxLen}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Apply enforces the ceiling on a single value.
// If WithReject is set and the value exceeds the ceiling, ErrExceedsCeiling is returned.
// Otherwise the value is silently truncated to maxLen bytes.
func (c *Ceiling) Apply(value string) (string, error) {
	if len(value) <= c.maxLen {
		return value, nil
	}
	if c.reject {
		return "", fmt.Errorf("%w: length %d exceeds max %d", ErrExceedsCeiling, len(value), c.maxLen)
	}
	return value[:c.maxLen], nil
}

// ApplyAll enforces the ceiling on every value in the map.
// Returns the first error encountered, leaving remaining keys unprocessed.
func (c *Ceiling) ApplyAll(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		applied, err := c.Apply(v)
		if err != nil {
			return nil, fmt.Errorf("ceiling: key %q: %w", k, err)
		}
		out[k] = applied
	}
	return out, nil
}
