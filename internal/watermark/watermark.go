// Package watermark embeds a hidden marker into secret values to track
// their origin and detect unauthorized copies or leaks.
package watermark

import (
	"errors"
	"fmt"
	"strings"
)

// Marker holds the watermark configuration.
type Marker struct {
	prefix    string
	separator string
}

// Option configures a Marker.
type Option func(*Marker)

// WithSeparator sets the separator between the watermark and the value.
func WithSeparator(sep string) Option {
	return func(m *Marker) {
		if sep != "" {
			m.separator = sep
		}
	}
}

// New creates a Marker with the given prefix namespace.
func New(prefix string, opts ...Option) (*Marker, error) {
	if prefix == "" {
		return nil, errors.New("watermark: prefix must not be empty")
	}
	m := &Marker{
		prefix:    prefix,
		separator: ":",
	}
	for _, o := range opts {
		o(m)
	}
	return m, nil
}

// Stamp embeds a watermark tag into the value using the given label.
func (m *Marker) Stamp(label, value string) string {
	tag := fmt.Sprintf("[wm%s%s%s%s]", m.separator, m.prefix, m.separator, label)
	return value + tag
}

// Strip removes the watermark tag from the value, returning the clean value
// and the embedded label. Returns an error if no watermark is found.
func (m *Marker) Strip(value string) (clean, label string, err error) {
	start := strings.LastIndex(value, "[wm")
	if start == -1 {
		return "", "", errors.New("watermark: no watermark found in value")
	}
	end := strings.Index(value[start:], "]")
	if end == -1 {
		return "", "", errors.New("watermark: malformed watermark tag")
	}
	tag := value[start : start+end+1]
	inner := tag[3 : len(tag)-1] // strip [wm and ]
	parts := strings.SplitN(inner, m.separator, 3)
	if len(parts) != 3 {
		return "", "", errors.New("watermark: invalid watermark format")
	}
	if parts[1] != m.prefix {
		return "", "", fmt.Errorf("watermark: prefix mismatch: got %q want %q", parts[1], m.prefix)
	}
	return value[:start], parts[2], nil
}

// Contains reports whether the value carries a watermark from this Marker.
func (m *Marker) Contains(value string) bool {
	_, _, err := m.Strip(value)
	return err == nil
}
