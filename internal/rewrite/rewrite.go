// Package rewrite provides key and value rewriting rules for secrets
// as they move between providers during sync or rotation operations.
package rewrite

import (
	"fmt"
	"strings"
)

// Rule defines a transformation applied to a secret key or value.
type Rule func(key, value string) (newKey, newValue string, err error)

// Rewriter applies an ordered list of rules to secrets.
type Rewriter struct {
	rules []Rule
}

// New returns a Rewriter with the given rules applied in order.
func New(rules ...Rule) *Rewriter {
	return &Rewriter{rules: rules}
}

// Apply runs all rules against the given key/value pair.
// Rules are applied sequentially; the output of each becomes the input of the next.
func (r *Rewriter) Apply(key, value string) (string, string, error) {
	k, v := key, value
	for _, rule := range r.rules {
		var err error
		k, v, err = rule(k, v)
		if err != nil {
			return "", "", fmt.Errorf("rewrite rule failed for key %q: %w", key, err)
		}
	}
	return k, v, nil
}

// ApplyAll rewrites all entries in the map, returning a new map.
func (r *Rewriter) ApplyAll(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		nk, nv, err := r.Apply(k, v)
		if err != nil {
			return nil, err
		}
		out[nk] = nv
	}
	return out, nil
}

// ReplaceKeyPrefix returns a Rule that replaces oldPrefix with newPrefix in keys.
func ReplaceKeyPrefix(oldPrefix, newPrefix string) Rule {
	return func(key, value string) (string, string, error) {
		if strings.HasPrefix(key, oldPrefix) {
			return newPrefix + strings.TrimPrefix(key, oldPrefix), value, nil
		}
		return key, value, nil
	}
}

// UpperCaseKey returns a Rule that converts the key to upper case.
func UpperCaseKey() Rule {
	return func(key, value string) (string, string, error) {
		return strings.ToUpper(key), value, nil
	}
}

// LowerCaseKey returns a Rule that converts the key to lower case.
func LowerCaseKey() Rule {
	return func(key, value string) (string, string, error) {
		return strings.ToLower(key), value, nil
	}
}

// AppendKeySuffix returns a Rule that appends a suffix to every key.
func AppendKeySuffix(suffix string) Rule {
	return func(key, value string) (string, string, error) {
		return key + suffix, value, nil
	}
}

// TrimValueSpace returns a Rule that trims whitespace from values.
func TrimValueSpace() Rule {
	return func(key, value string) (string, string, error) {
		return key, strings.TrimSpace(value), nil
	}
}
