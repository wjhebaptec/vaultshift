// Package stamp attaches and reads metadata timestamps on secret values.
package stamp

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const defaultSep = "|"

// Stamper embeds a timestamp into a secret value and can later extract it.
type Stamper struct {
	sep string
	clock func() time.Time
}

// Option configures a Stamper.
type Option func(*Stamper)

// WithSeparator sets the separator used between value and timestamp.
func WithSeparator(sep string) Option {
	return func(s *Stamper) { s.sep = sep }
}

// WithClock sets a custom clock for testing.
func WithClock(fn func() time.Time) Option {
	return func(s *Stamper) { s.clock = fn }
}

// New returns a new Stamper.
func New(opts ...Option) (*Stamper, error) {
	s := &Stamper{sep: defaultSep, clock: time.Now}
	for _, o := range opts {
		o(s)
	}
	if s.sep == "" {
		return nil, errors.New("stamp: separator must not be empty")
	}
	return s, nil
}

// Attach appends a UTC timestamp to value.
func (s *Stamper) Attach(value string) string {
	ts := s.clock().UTC().Format(time.RFC3339)
	return fmt.Sprintf("%s%s%s", value, s.sep, ts)
}

// Extract splits a stamped value into its original value and timestamp.
func (s *Stamper) Extract(stamped string) (value string, ts time.Time, err error) {
	idx := strings.LastIndex(stamped, s.sep)
	if idx < 0 {
		return "", time.Time{}, errors.New("stamp: no timestamp found in value")
	}
	rawTS := stamped[idx+len(s.sep):]
	ts, err = time.Parse(time.RFC3339, rawTS)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("stamp: invalid timestamp %q: %w", rawTS, err)
	}
	return stamped[:idx], ts, nil
}

// Age returns how long ago the stamped value was created.
func (s *Stamper) Age(stamped string) (time.Duration, error) {
	_, ts, err := s.Extract(stamped)
	if err != nil {
		return 0, err
	}
	return s.clock().UTC().Sub(ts), nil
}
