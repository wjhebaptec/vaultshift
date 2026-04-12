// Package shadow implements shadow-write semantics for secret providers.
//
// A Shadow wraps two providers — a primary and a shadow — and transparently
// dual-writes every Put operation to both. Reads always return the primary
// value, keeping production behaviour unchanged.
//
// Two modes are supported:
//
//   - ModeWriteOnly: writes to both; reads only from primary.
//   - ModeCompare:   writes to both; on reads, compares primary and shadow
//     values and records any divergence in an in-memory mismatch log.
//
// Mismatches can be inspected via Mismatches() and cleared with
// ResetMismatches(). This is useful for validating a new provider before
// cutting over traffic.
//
// Once the shadow is confirmed correct, Promote() copies all shadow keys
// into the primary provider, completing the migration.
package shadow
