// Package sync provides functionality for synchronizing secrets across multiple
// secret management providers.
//
// The sync package enables users to:
//   - Sync individual secrets from a source provider to one or more target providers
//   - Sync all secrets from a source to targets
//   - Sync secrets selectively using filter functions
//
// Example usage:
//
//	registry := provider.NewRegistry()
//	registry.Register("aws", awsProvider)
//	registry.Register("gcp", gcpProvider)
//
//	syncer := sync.New(registry)
//
//	// Sync a single secret
//	err := syncer.SyncSecret(ctx, "api-key", "aws", []string{"gcp"})
//
//	// Sync all secrets
//	err = syncer.SyncAll(ctx, "aws", []string{"gcp"})
//
//	// Sync with filter
//	filter := sync.NewPrefixFilter("prod/")
//	err = syncer.SyncWithFilter(ctx, "aws", []string{"gcp"}, filter)
//
// Filters:
//
// The package provides several built-in filter functions:
//   - PrefixFilter: Match secrets with a specific prefix
//   - SuffixFilter: Match secrets with a specific suffix
//   - RegexFilter: Match secrets against a regex pattern
//   - IncludeFilter: Only include specific secret keys
//   - ExcludeFilter: Exclude specific secret keys
//   - CombineFilters: Combine multiple filters with AND logic
//   - AnyFilter: Combine multiple filters with OR logic
//
// Custom filters can be created by implementing the FilterFunc type:
//
//	type FilterFunc func(secretKey string) bool
package sync
