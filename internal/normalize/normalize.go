// Package normalize provides key normalization utilities for secret names
// across different cloud provider naming conventions.
package normalize

import (
	"strings"
	"unicode"
)

// Style defines the target naming convention.
type Style int

const (
	// StyleSnake converts keys to snake_case (e.g. "mySecret" -> "my_secret").
	StyleSnake Style = iota
	// StyleKebab converts keys to kebab-case (e.g. "mySecret" -> "my-secret").
	StyleKebab
	// StyleScream converts keys to SCREAMING_SNAKE_CASE.
	StyleScream
	// StyleDot converts keys to dot.notation.
	StyleDot
)

// Normalizer converts secret key names to a consistent style.
type Normalizer struct {
	style  Style
	prefix string
}

// Option configures a Normalizer.
type Option func(*Normalizer)

// WithStyle sets the target naming style.
func WithStyle(s Style) Option {
	return func(n *Normalizer) { n.style = s }
}

// WithPrefix prepends a fixed prefix after normalization.
func WithPrefix(p string) Option {
	return func(n *Normalizer) { n.prefix = p }
}

// New creates a Normalizer with the provided options.
func New(opts ...Option) *Normalizer {
	n := &Normalizer{style: StyleSnake}
	for _, o := range opts {
		o(n)
	}
	return n
}

// Normalize converts a single key to the configured style.
func (n *Normalizer) Normalize(key string) string {
	words := split(key)
	var result string
	switch n.style {
	case StyleKebab:
		result = strings.Join(lower(words), "-")
	case StyleScream:
		result = strings.Join(upper(words), "_")
	case StyleDot:
		result = strings.Join(lower(words), ".")
	default: // StyleSnake
		result = strings.Join(lower(words), "_")
	}
	if n.prefix != "" {
		return n.prefix + result
	}
	return result
}

// NormalizeAll applies Normalize to every key in the map, returning a new map.
func (n *Normalizer) NormalizeAll(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[n.Normalize(k)] = v
	}
	return out
}

// split breaks a key into words by camelCase, snake_case, kebab-case, or dots.
func split(key string) []string {
	var words []string
	var cur strings.Builder
	for i, r := range key {
		if r == '_' || r == '-' || r == '.' || r == '/' {
			if cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
			}
			continue
		}
		if i > 0 && unicode.IsUpper(r) && cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	return words
}

func lower(words []string) []string {
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = strings.ToLower(w)
	}
	return out
}

func upper(words []string) []string {
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = strings.ToUpper(w)
	}
	return out
}
