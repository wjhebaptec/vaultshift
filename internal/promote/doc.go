// Package promote provides functionality for promoting secrets between
// environments (e.g. staging → production) via the vaultshift provider
// registry.
//
// # Overview
//
// A Promoter reads secrets from a source provider and writes them to a
// destination provider. An optional KeyTransform function can rename keys
// during the copy (e.g. prepend an environment prefix). A DryRun mode
// reports what would be written without making any changes.
//
// # Usage
//
//	reg := provider.NewRegistry()
//	reg.Register("staging", stagingProvider)
//	reg.Register("prod",    prodProvider)
//
//	p := promote.New(reg, promote.Options{
//		KeyTransform: func(k string) string { return "prod/" + k },
//	})
//
//	// Promote a single secret
//	r := p.Promote(ctx, "staging", "prod", "DB_PASSWORD")
//
//	// Promote all secrets
//	results := p.PromoteAll(ctx, "staging", "prod")
//	if promote.HasFailures(results) { ... }
package promote
