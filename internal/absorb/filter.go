package absorb

import "strings"

// Filter decides whether a key should be absorbed.
type Filter func(key string) bool

// WithSourcePrefix returns a Filter that passes only keys starting with prefix.
func WithSourcePrefix(prefix string) Filter {
	return func(key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

// WithSourceSuffix returns a Filter that passes only keys ending with suffix.
func WithSourceSuffix(suffix string) Filter {
	return func(key string) bool {
		return strings.HasSuffix(key, suffix)
	}
}

// WithExclude returns a Filter that rejects keys matching any of the given substrings.
func WithExclude(substrings ...string) Filter {
	return func(key string) bool {
		for _, s := range substrings {
			if strings.Contains(key, s) {
				return false
			}
		}
		return true
	}
}

// ChainFilters combines multiple filters; all must pass for a key to be accepted.
func ChainFilters(filters ...Filter) Filter {
	return func(key string) bool {
		for _, f := range filters {
			if !f(key) {
				return false
			}
		}
		return true
	}
}

// Allowed returns true when f is nil (accept all) or f(key) is true.
func Allowed(f Filter, key string) bool {
	if f == nil {
		return true
	}
	return f(key)
}
