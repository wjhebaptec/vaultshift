// Package prefix provides namespace-aware key qualification for secret
// providers.
//
// # Overview
//
// A Prefixer wraps raw secret keys with a configurable namespace and
// separator so that multiple tenants or environments can share the same
// underlying provider without key collisions.
//
// # Usage
//
//	pfx, err := prefix.New("prod", prefix.WithSeparator("/"))
//	w, err   := prefix.Wrap(myProvider, pfx)
//
//	// Internally stored as "prod/db-password"
//	_ = w.Put(ctx, "db-password", value)
//
// The Wrapped provider implements provider.Provider so it can be used
// anywhere a plain provider is expected.
package prefix
