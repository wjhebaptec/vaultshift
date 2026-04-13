// Package clamp enforces minimum and maximum bounds on secret values.
// Values that fall outside the configured range are either clamped to the
// nearest bound or rejected, depending on the configured mode.
package clamp

import (
	"errors"
	"fmt"
)

// ErrOutOfRange is returned when a value exceeds bounds and reject mode is on.
var ErrOutOfRange = errors.New("clamp: value length out of range")

// Option configures a Clamper.
type Option func(*Clamper)

// WithReject causes Apply to return an error instead of clamping the value.
func WithReject() Option {
	return func(c *Clamper) { c.reject = true }
}

// Clamper enforces min/max length bounds on string values.
type Clamper struct {
	min    int
	max    int
	reject bool
}

// New creates a Clamper with the given minimum and maximum lengths.
// min must be >= 0 and max must be >= min.
func New(min, max int, opts ...Option) (*Clamper, error) {
	if min < 0 {
		return nil, fmt.Errorf("clamp: min must be >= 0, got %d", min)
	}
	if max < min {
		return nil, fmt.Errorf("clamp: max (%d) must be >= min (%d)", max, min)
	}
	c := &Clamper{min: min, max: max}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Apply enforces the bounds on v. If reject mode is enabled and the value
// falls outside the range, ErrOutOfRange is returned. Otherwise the value
// is padded with spaces on the right (if too short) or truncated (if too long).
func (c *Clamper) Apply(v string) (string, error) {
	l := len(v)
	if l >= c.min && l <= c.max {
		return v, nil
	}
	if c.reject {
		return "", fmt.Errorf("%w: length %d not in [%d, %d]", ErrOutOfRange, l, c.min, c.max)
	}
	if l < c.min {
		for len(v) < c.min {
			v += " "
		}
		return v, nil
	}
	// l > c.max
	return v[:c.max], nil
}

// ApplyAll applies the clamper to every value in the map, returning the
// first error encountered. The returned map contains only successfully
// clamped entries up to the point of failure.
func (c *Clamper) ApplyAll(m map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(m))
	for k, v := range m {
		clamped, err := c.Apply(v)
		if err != nil {
			return out, fmt.Errorf("clamp: key %q: %w", k, err)
		}
		out[k] = clamped
	}
	return out, nil
}
