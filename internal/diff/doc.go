// Package diff provides utilities for comparing secrets across providers.
//
// Use Compare to produce a DiffResult describing which secrets were added,
// removed, updated, or unchanged between two snapshots. Use HasDrift to
// quickly determine whether any discrepancy exists.
package diff
