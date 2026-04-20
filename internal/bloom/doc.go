// Package bloom implements a space-efficient probabilistic data structure
// for testing whether a secret key is a member of a set.
//
// A bloom filter can definitively report that a key has NOT been added,
// but may produce false positives — reporting a key as present when it
// has not been added. The false-positive rate decreases as the filter
// size and number of hash functions increase.
//
// Typical usage:
//
//	f, err := bloom.New(bloom.WithSize(2048), bloom.WithHashFunctions(4))
//	if err != nil { ... }
//
//	f.Add("prod/db/password")
//
//	if f.MayContain("prod/db/password") {
//		// probably present — do a real lookup
//	}
//
// The filter is safe for concurrent use.
package bloom
