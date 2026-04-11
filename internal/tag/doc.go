// Package tag provides a thread-safe in-memory store for attaching arbitrary
// string key/value metadata tags to secret keys managed by vaultshift.
//
// Tags can be used to:
//   - Group secrets by environment (e.g. env=production)
//   - Mark secrets for selective rotation or sync
//   - Drive policy decisions based on label presence
//
// Example:
//
//	s := tag.New()
//	s.Set("db/password", "env", "production")
//	s.Set("db/password", "team", "platform")
//
//	matches := s.MatchAll(map[string]string{"env": "production"})
//	// matches == ["db/password"]
package tag
