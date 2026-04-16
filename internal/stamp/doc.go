// Package stamp provides utilities for embedding and extracting UTC timestamps
// from secret values. This is useful for tracking when a secret was last
// written or rotated without relying on external metadata.
//
// Usage:
//
//	s, _ := stamp.New()
//	stamped := s.Attach("my-secret-value")
//	value, ts, err := s.Extract(stamped)
//	age, err := s.Age(stamped)
package stamp
