// Package redact provides utilities for redacting sensitive secret values
// from log output, error messages, and structured data before they are
// written to any external sink.
package redact

import (
	"strings"
	"sync"
)

// Redactor holds a registry of known sensitive values and replaces them
// with a configurable placeholder string.
type Redactor struct {
	mu          sync.RWMutex
	sensitive   map[string]struct{}
	placeholder string
}

// Option is a functional option for configuring a Redactor.
type Option func(*Redactor)

// WithPlaceholder sets the string used to replace sensitive values.
// Defaults to "[REDACTED]".
func WithPlaceholder(p string) Option {
	return func(r *Redactor) {
		if p != "" {
			r.placeholder = p
		}
	}
}

// New creates a new Redactor with the given options.
func New(opts ...Option) *Redactor {
	r := &Redactor{
		sensitive:   make(map[string]struct{}),
		placeholder: "[REDACTED]",
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Register marks value as sensitive. Future calls to Redact will replace
// any occurrence of value with the configured placeholder.
func (r *Redactor) Register(value string) {
	if value == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sensitive[value] = struct{}{}
}

// Forget removes value from the sensitive registry.
func (r *Redactor) Forget(value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sensitive, value)
}

// Redact replaces all registered sensitive values found in input with
// the configured placeholder and returns the sanitised string.
func (r *Redactor) Redact(input string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for secret := range r.sensitive {
		input = strings.ReplaceAll(input, secret, r.placeholder)
	}
	return input
}

// RedactMap returns a copy of m where every value has been redacted.
// Keys are left untouched.
func (r *Redactor) RedactMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = r.Redact(v)
	}
	return out
}
