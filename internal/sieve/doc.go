// Package sieve implements a composable key-filtering mechanism for use in
// secret pipelines. A Sieve is constructed from one or more Rules; a key is
// allowed through only when every Rule returns true.
//
// Built-in rules:
//
//	MatchPrefix(p)  – key must start with p
//	MatchSuffix(s)  – key must end with s
//	MatchRegex(re)  – key must match the regular expression
//	Deny(rule)      – inverts any existing rule
//
// Example:
//
//	s, err := sieve.New(
//	    sieve.MatchPrefix("prod/"),
//	    sieve.Deny(sieve.MatchSuffix("_DEPRECATED")),
//	)
//	allowed := s.Filter(keys)
package sieve
