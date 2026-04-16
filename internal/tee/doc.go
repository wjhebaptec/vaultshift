// Package tee provides a provider wrapper that duplicates operations to a
// secondary provider. It is useful for warming up caches, shadowing traffic
// to a new backend, or keeping a read-replica in sync without changing
// upstream call sites.
//
// By default both reads and writes are mirrored. Use WithWriteOnly to
// restrict mirroring to Put and Delete operations only.
//
// Errors from the secondary provider are silently discarded so that tee
// never degrades the primary path.
package tee
