package relay

import (
	"context"
	"strings"
)

// RuleFilter is a predicate that decides whether a rule should be forwarded.
type RuleFilter func(rule Rule) bool

// WithSourcePrefix returns a RuleFilter that only passes rules whose source
// key starts with the given prefix.
func WithSourcePrefix(prefix string) RuleFilter {
	return func(r Rule) bool {
		return strings.HasPrefix(r.SourceKey, prefix)
	}
}

// WithDestProvider returns a RuleFilter that only passes rules targeting the
// named destination provider.
func WithDestProvider(name string) RuleFilter {
	return func(r Rule) bool {
		return r.DestProvider == name
	}
}

// ChainFilters combines multiple RuleFilters with AND semantics.
func ChainFilters(filters ...RuleFilter) RuleFilter {
	return func(r Rule) bool {
		for _, f := range filters {
			if !f(r) {
				return false
			}
		}
		return true
	}
}

// FilteredRelay wraps a Relay and only forwards rules that pass the filter.
type FilteredRelay struct {
	base   *Relay
	filter RuleFilter
}

// NewFiltered creates a FilteredRelay that only forwards rules accepted by f.
func NewFiltered(r *Relay, f RuleFilter) *FilteredRelay {
	return &FilteredRelay{base: r, filter: f}
}

// Forward applies only the rules that pass the configured filter.
func (f *FilteredRelay) Forward(ctx context.Context) []Result {
	matched := make([]Rule, 0, len(f.base.rules))
	for _, rule := range f.base.rules {
		if f.filter(rule) {
			matched = append(matched, rule)
		}
	}
	orig := f.base.rules
	f.base.rules = matched
	results := f.base.Forward(ctx)
	f.base.rules = orig
	return results
}
