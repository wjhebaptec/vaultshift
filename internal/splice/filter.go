package splice

import "strings"

// Filter decides whether a (sourceProvider, key) pair should be spliced.
type Filter func(sourceProvider, key string) bool

// WithSourcePrefix returns a Filter that only allows keys with the given prefix.
func WithSourcePrefix(prefix string) Filter {
	return func(_, key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

// WithSourceSuffix returns a Filter that only allows keys with the given suffix.
func WithSourceSuffix(suffix string) Filter {
	return func(_, key string) bool {
		return strings.HasSuffix(key, suffix)
	}
}

// WithProviderName returns a Filter that only allows the named source provider.
func WithProviderName(name string) Filter {
	return func(sourceProvider, _ string) bool {
		return sourceProvider == name
	}
}

// ChainFilters combines multiple filters with AND semantics.
// An empty chain always returns true.
func ChainFilters(filters ...Filter) Filter {
	return func(sourceProvider, key string) bool {
		for _, f := range filters {
			if !f(sourceProvider, key) {
				return false
			}
		}
		return true
	}
}

// Allowed is a convenience wrapper that applies a Filter (which may be nil).
func Allowed(f Filter, sourceProvider, key string) bool {
	if f == nil {
		return true
	}
	return f(sourceProvider, key)
}
