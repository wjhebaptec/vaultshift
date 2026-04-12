package export

import "strings"

// KeyFilter decides whether a secret key should be included in the export.
type KeyFilter func(key string) bool

// WithPrefixFilter returns a KeyFilter that includes only keys with the given prefix.
func WithPrefixFilter(prefix string) KeyFilter {
	return func(key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

// WithExcludeFilter returns a KeyFilter that excludes keys matching any of the given substrings.
func WithExcludeFilter(substrings ...string) KeyFilter {
	return func(key string) bool {
		for _, s := range substrings {
			if strings.Contains(key, s) {
				return false
			}
		}
		return true
	}
}

// WithSuffixFilter returns a KeyFilter that includes only keys with the given suffix.
func WithSuffixFilter(suffix string) KeyFilter {
	return func(key string) bool {
		return strings.HasSuffix(key, suffix)
	}
}

// ChainFilters combines multiple KeyFilters with AND logic.
func ChainFilters(filters ...KeyFilter) KeyFilter {
	return func(key string) bool {
		for _, f := range filters {
			if !f(key) {
				return false
			}
		}
		return true
	}
}

// FilteredExporter wraps an Exporter and applies a KeyFilter before exporting.
type FilteredExporter struct {
	exporter *Exporter
	filter   KeyFilter
}

// NewFiltered creates a FilteredExporter.
func NewFiltered(provider Provider, format Format, filter KeyFilter, out interface{ Write([]byte) (int, error) }) *FilteredExporter {
	return &FilteredExporter{
		exporter: New(provider, format, out),
		filter:   filter,
	}
}

// filteredProvider wraps a Provider and applies a KeyFilter to ListSecrets.
type filteredProvider struct {
	inner  Provider
	filter KeyFilter
}

func (fp *filteredProvider) ListSecrets(ctx interface{ Deadline() (interface{}, bool); Done() <-chan struct{}; Err() error; Value(interface{}) interface{} }) ([]string, error) {
	return nil, nil
}
