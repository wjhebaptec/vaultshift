// Package label provides key-value labelling for secrets managed by vaultshift.
//
// Labels are arbitrary string metadata attached to a (provider, secretKey) pair.
// They are stored in-memory and are not persisted to any backend provider.
//
// Typical uses include:
//   - Grouping secrets by environment (env=prod, env=staging)
//   - Marking secrets by owning team (team=platform)
//   - Filtering secrets for targeted rotation or sync operations
//
// Example:
//
//	mgr := label.New()
//	_ = mgr.Set("aws", "db/password", label.Labels{"env": "prod"})
//
//	matched := mgr.Filter("aws", allKeys, label.Labels{"env": "prod"})
//	for _, key := range matched {
//		// rotate or sync only prod secrets
//	}
package label
