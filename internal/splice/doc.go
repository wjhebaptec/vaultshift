// Package splice provides a Splicer that merges secrets from one or more source
// providers into a single destination provider.
//
// Basic usage:
//
//	s, err := splice.New(registry, "vault", nil)
//	if err != nil { ... }
//	// Copy a single key
//	if err := s.Splice(ctx, "aws", "prod/db/password"); err != nil { ... }
//	// Copy many keys
//	s.SpliceAll(ctx, "gcp", []string{"api/key", "api/secret"})
//	if s.HasFailures() { /* inspect s.Results() */ }
//
// Key rewriting:
//
//	rewrite := func(src, key string) string {
//		return src + "/" + key   // namespace by source
//	}
//	s, _ := splice.New(registry, "vault", rewrite)
//
// Filtering (using the filter sub-helpers):
//
//	f := splice.ChainFilters(
//		splice.WithSourcePrefix("prod/"),
//		splice.WithProviderName("aws"),
//	)
//	if splice.Allowed(f, "aws", key) {
//		s.Splice(ctx, "aws", key)
//	}
package splice
