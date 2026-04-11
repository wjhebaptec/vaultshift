// Package transform provides value transformation pipelines for secrets,
// allowing chained mutations such as trimming, casing, encoding, and prefixing.
package transform

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Func is a function that transforms a secret value.
type Func func(value string) (string, error)

// Transformer applies a chain of Func transformations to secret values.
type Transformer struct {
	steps []Func
}

// New returns a Transformer with the given transformation steps.
func New(steps ...Func) *Transformer {
	return &Transformer{steps: steps}
}

// Apply runs all transformation steps against value in order.
// It returns the transformed result or the first error encountered.
func (t *Transformer) Apply(value string) (string, error) {
	result := value
	for i, step := range t.steps {
		var err error
		result, err = step(result)
		if err != nil {
			return "", fmt.Errorf("transform step %d: %w", i, err)
		}
	}
	return result, nil
}

// ApplyAll applies the transformer to a map of secret values.
// It returns a new map with transformed values or the first error encountered.
func (t *Transformer) ApplyAll(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		transformed, err := t.Apply(v)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", k, err)
		}
		out[k] = transformed
	}
	return out, nil
}

// TrimSpace returns a Func that trims leading and trailing whitespace.
func TrimSpace() Func {
	return func(value string) (string, error) {
		return strings.TrimSpace(value), nil
	}
}

// ToUpper returns a Func that converts the value to uppercase.
func ToUpper() Func {
	return func(value string) (string, error) {
		return strings.ToUpper(value), nil
	}
}

// ToLower returns a Func that converts the value to lowercase.
func ToLower() Func {
	return func(value string) (string, error) {
		return strings.ToLower(value), nil
	}
}

// AddPrefix returns a Func that prepends prefix to the value.
func AddPrefix(prefix string) Func {
	return func(value string) (string, error) {
		return prefix + value, nil
	}
}

// Base64Encode returns a Func that base64-encodes the value.
func Base64Encode() Func {
	return func(value string) (string, error) {
		return base64.StdEncoding.EncodeToString([]byte(value)), nil
	}
}

// Base64Decode returns a Func that base64-decodes the value.
func Base64Decode() Func {
	return func(value string) (string, error) {
		b, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return "", fmt.Errorf("base64 decode: %w", err)
		}
		return string(b), nil
	}
}
