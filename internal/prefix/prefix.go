// Package prefix provides utilities for managing key prefix namespacing
// across secret providers, enabling scoped access and isolation between
// environments or teams.
package prefix

import (
	"fmt"
	"strings"
)

// Prefixer wraps and unwraps secret keys with a configured namespace prefix.
type Prefixer struct {
	prefix    string
	separator string
}

// Option configures a Prefixer.
type Option func(*Prefixer)

// WithSeparator sets the separator between prefix and key (default: "/").
func WithSeparator(sep string) Option {
	return func(p *Prefixer) {
		p.separator = sep
	}
}

// New creates a Prefixer with the given namespace prefix.
func New(namespace string, opts ...Option) (*Prefixer, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, fmt.Errorf("prefix: namespace must not be empty")
	}
	p := &Prefixer{
		prefix:    namespace,
		separator: "/",
	}
	for _, o := range opts {
		o(p)
	}
	return p, nil
}

// Wrap prepends the namespace prefix to the given key.
func (p *Prefixer) Wrap(key string) string {
	if key == "" {
		return p.prefix
	}
	return p.prefix + p.separator + key
}

// Unwrap removes the namespace prefix from the given key.
// Returns the stripped key and true if the prefix was present,
// or the original key and false if not.
func (p *Prefixer) Unwrap(key string) (string, bool) {
	expected := p.prefix + p.separator
	if strings.HasPrefix(key, expected) {
		return strings.TrimPrefix(key, expected), true
	}
	if key == p.prefix {
		return "", true
	}
	return key, false
}

// WrapAll applies Wrap to every key in the provided map, returning a new map.
func (p *Prefixer) WrapAll(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[p.Wrap(k)] = v
	}
	return out
}

// UnwrapAll removes the namespace prefix from every key in the map.
// Keys that do not carry the prefix are omitted from the result.
func (p *Prefixer) UnwrapAll(secrets map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range secrets {
		if stripped, ok := p.Unwrap(k); ok {
			out[stripped] = v
		}
	}
	return out
}

// Namespace returns the configured prefix string.
func (p *Prefixer) Namespace() string {
	return p.prefix
}
