// Package environ provides utilities for rendering secrets into
// environment-variable maps, suitable for process injection or .env export.
package environ

import (
	"context"
	"fmt"
	"strings"

	"github.com/vaultshift/internal/provider"
)

// Option configures an Exporter.
type Option func(*Exporter)

// WithUpperCase converts all keys to UPPER_CASE before writing.
func WithUpperCase() Option {
	return func(e *Exporter) { e.upperCase = true }
}

// WithPrefix prepends a fixed string to every key.
func WithPrefix(p string) Option {
	return func(e *Exporter) { e.prefix = p }
}

// WithQuoteValues wraps values in double-quotes in the rendered output.
func WithQuoteValues() Option {
	return func(e *Exporter) { e.quoteValues = true }
}

// Exporter renders secrets from a provider into an environment-variable map.
type Exporter struct {
	reg         *provider.Registry
	upperCase   bool
	prefix      string
	quoteValues bool
}

// New creates an Exporter backed by the given registry.
func New(reg *provider.Registry, opts ...Option) *Exporter {
	e := &Exporter{reg: reg}
	for _, o := range opts {
		o(e)
	}
	return e
}

// Render lists all secrets from the named provider and returns a map of
// environment-variable key → value strings.
func (e *Exporter) Render(ctx context.Context, providerName string) (map[string]string, error) {
	p, ok := e.reg.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("environ: unknown provider %q", providerName)
	}

	keys, err := p.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("environ: list secrets: %w", err)
	}

	out := make(map[string]string, len(keys))
	for _, k := range keys {
		val, err := p.GetSecret(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("environ: get %q: %w", k, err)
		}
		envKey := e.formatKey(k)
		out[envKey] = e.formatValue(val)
	}
	return out, nil
}

// Lines returns the map as a slice of KEY=VALUE strings, sorted by key.
func (e *Exporter) Lines(ctx context.Context, providerName string) ([]string, error) {
	m, err := e.Render(ctx, providerName)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(m))
	for k, v := range m {
		lines = append(lines, k+"="+v)
	}
	return lines, nil
}

func (e *Exporter) formatKey(k string) string {
	if e.prefix != "" {
		k = e.prefix + k
	}
	if e.upperCase {
		k = strings.ToUpper(k)
	}
	return k
}

func (e *Exporter) formatValue(v string) string {
	if e.quoteValues {
		return `"` + v + `"`
	}
	return v
}
