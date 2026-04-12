// Package rewrite provides composable key and value rewriting rules
// for use during secret sync and rotation operations in vaultshift.
//
// Rules are plain functions with the signature:
//
//	func(key, value string) (newKey, newValue string, err error)
//
// A Rewriter chains multiple rules together, applying them in order.
// Built-in rules cover common transformations such as prefix replacement,
// case conversion, suffix appending, and value whitespace trimming.
//
// Example usage:
//
//	r := rewrite.New(
//	    rewrite.ReplaceKeyPrefix("prod/", "staging/"),
//	    rewrite.UpperCaseKey(),
//	    rewrite.TrimValueSpace(),
//	)
//	newKey, newVal, err := r.Apply("prod/db_pass", "  secret  ")
package rewrite
