// Package policy defines rotation and sync policies for secrets,
// including TTL enforcement, allowed key patterns, and required
// provider targets.
package policy

import (
	"errors"
	"regexp"
	"time"
)

// Policy defines rules governing how secrets are rotated and synced.
type Policy struct {
	// Name is a human-readable identifier for this policy.
	Name string
	// MaxAge is the maximum allowed age of a secret before rotation is required.
	MaxAge time.Duration
	// AllowedPattern restricts which secret keys this policy applies to.
	AllowedPattern *regexp.Regexp
	// RequiredTargets lists provider names that must receive every sync.
	RequiredTargets []string
}

// Option is a functional option for configuring a Policy.
type Option func(*Policy)

// WithMaxAge sets the maximum age before a secret must be rotated.
func WithMaxAge(d time.Duration) Option {
	return func(p *Policy) {
		p.MaxAge = d
	}
}

// WithAllowedPattern restricts the policy to keys matching the given regex.
func WithAllowedPattern(pattern string) Option {
	return func(p *Policy) {
		p.AllowedPattern = regexp.MustCompile(pattern)
	}
}

// WithRequiredTargets specifies providers that must be synced.
func WithRequiredTargets(targets ...string) Option {
	return func(p *Policy) {
		p.RequiredTargets = targets
	}
}

// New creates a Policy with the given name and options.
func New(name string, opts ...Option) *Policy {
	p := &Policy{Name: name}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Validate checks whether a secret key and its last-rotated time comply
// with this policy. It returns an error describing the first violation found.
func (p *Policy) Validate(key string, lastRotated time.Time) error {
	if p.AllowedPattern != nil && !p.AllowedPattern.MatchString(key) {
		return errors.New("policy: key \"" + key + "\" does not match allowed pattern " + p.AllowedPattern.String())
	}
	if p.MaxAge > 0 && !lastRotated.IsZero() {
		age := time.Since(lastRotated)
		if age > p.MaxAge {
			return errors.New("policy: secret \"" + key + "\" exceeds max age")
		}
	}
	return nil
}
