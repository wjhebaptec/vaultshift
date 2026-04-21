// Package scrub provides a middleware that removes or replaces secret values
// matching a set of registered patterns before they are written to a provider.
package scrub

import (
	"context"
	"errors"
	"regexp"
	"sync"
)

// Rule describes a single scrubbing rule.
type Rule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// Scrubber applies registered rules to secret values.
type Scrubber struct {
	mu    sync.RWMutex
	rules []Rule
}

// New returns a new Scrubber with no rules.
func New() *Scrubber {
	return &Scrubber{}
}

// AddRule registers a pattern and the replacement string to use when the
// pattern matches a secret value. An empty replacement effectively redacts
// the matched portion.
func (s *Scrubber) AddRule(pattern string, replacement string) error {
	if pattern == "" {
		return errors.New("scrub: pattern must not be empty")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, Rule{Pattern: re, Replacement: replacement})
	return nil
}

// Clean applies all registered rules to value and returns the result.
func (s *Scrubber) Clean(_ context.Context, value string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rules {
		value = r.Pattern.ReplaceAllString(value, r.Replacement)
	}
	return value
}

// Rules returns a snapshot of the currently registered rules.
func (s *Scrubber) Rules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Rule, len(s.rules))
	copy(out, s.rules)
	return out
}
