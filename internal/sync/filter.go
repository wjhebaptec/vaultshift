package sync

import (
	"regexp"
	"strings"
)

// FilterFunc is a function that determines if a secret should be synced
type FilterFunc func(secretKey string) bool

// NewPrefixFilter creates a filter that matches secrets with a given prefix
func NewPrefixFilter(prefix string) FilterFunc {
	return func(secretKey string) bool {
		return strings.HasPrefix(secretKey, prefix)
	}
}

// NewSuffixFilter creates a filter that matches secrets with a given suffix
func NewSuffixFilter(suffix string) FilterFunc {
	return func(secretKey string) bool {
		return strings.HasSuffix(secretKey, suffix)
	}
}

// NewRegexFilter creates a filter that matches secrets against a regex pattern
func NewRegexFilter(pattern string) (FilterFunc, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(secretKey string) bool {
		return re.MatchString(secretKey)
	}, nil
}

// NewExcludeFilter creates a filter that excludes specific secret keys
func NewExcludeFilter(excludeKeys []string) FilterFunc {
	excludeMap := make(map[string]bool)
	for _, key := range excludeKeys {
		excludeMap[key] = true
	}
	return func(secretKey string) bool {
		return !excludeMap[secretKey]
	}
}

// NewIncludeFilter creates a filter that only includes specific secret keys
func NewIncludeFilter(includeKeys []string) FilterFunc {
	includeMap := make(map[string]bool)
	for _, key := range includeKeys {
		includeMap[key] = true
	}
	return func(secretKey string) bool {
		return includeMap[secretKey]
	}
}

// CombineFilters combines multiple filters with AND logic
func CombineFilters(filters ...FilterFunc) FilterFunc {
	return func(secretKey string) bool {
		for _, filter := range filters {
			if !filter(secretKey) {
				return false
			}
		}
		return true
	}
}

// AnyFilter combines multiple filters with OR logic
func AnyFilter(filters ...FilterFunc) FilterFunc {
	return func(secretKey string) bool {
		for _, filter := range filters {
			if filter(secretKey) {
				return true
			}
		}
		return false
	}
}
