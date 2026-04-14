// Package sieve provides a filtering layer that selectively passes secrets
// through a pipeline based on configurable match rules.
package sieve

import (
	"errors"
	"regexp"
)

// Rule decides whether a given key should be allowed through.
type Rule func(key string) bool

// Sieve holds an ordered list of rules. A key passes when ALL rules return true.
type Sieve struct {
	rules []Rule
}

// New returns a Sieve with the provided rules.
func New(rules ...Rule) (*Sieve, error) {
	for i, r := range rules {
		if r == nil {
			return nil, errors.New("sieve: nil rule at index " + itoa(i))
		}
	}
	return &Sieve{rules: rules}, nil
}

// Allow returns true if the key satisfies every rule in the sieve.
func (s *Sieve) Allow(key string) bool {
	for _, r := range s.rules {
		if !r(key) {
			return false
		}
	}
	return true
}

// Filter returns only the keys from the input slice that pass the sieve.
func (s *Sieve) Filter(keys []string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if s.Allow(k) {
			out = append(out, k)
		}
	}
	return out
}

// MatchPrefix creates a Rule that accepts keys beginning with prefix.
func MatchPrefix(prefix string) Rule {
	return func(key string) bool {
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
}

// MatchSuffix creates a Rule that accepts keys ending with suffix.
func MatchSuffix(suffix string) Rule {
	return func(key string) bool {
		if len(key) < len(suffix) {
			return false
		}
		return key[len(key)-len(suffix):] == suffix
	}
}

// MatchRegex creates a Rule that accepts keys matching the compiled pattern.
func MatchRegex(pattern string) (Rule, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(key string) bool { return re.MatchString(key) }, nil
}

// Deny creates a Rule that rejects keys accepted by inner.
func Deny(inner Rule) Rule {
	return func(key string) bool { return !inner(key) }
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [10]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
