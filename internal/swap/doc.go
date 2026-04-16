// Package swap implements atomic secret swapping between two secret providers.
//
// A Swapper reads a secret from a source provider and writes it to a
// destination provider. When created with bidirectional=true, the
// destination's previous value is written back to the source, effectively
// exchanging the two values in a single operation.
//
// Example usage:
//
//	s, _ := swap.New(registry, false)
//	res := s.Swap(ctx, "aws", "gcp", "db/password")
//	if res.Err != nil {
//		log.Fatal(res.Err)
//	}
//
// SwapAll operates over a slice of keys and returns one Result per key.
// Use HasFailures to check whether any individual swap failed.
package swap
