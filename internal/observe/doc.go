// Package observe wires together metrics, audit logging, and event publishing
// into a single Observer that can be attached to any provider operation.
//
// Usage:
//
//	obs := observe.New("aws", collector, logger, bus)
//	obs.Record(ctx, observe.KindGet, "my/secret", err)
//
// Each call to Record will:
//   - Append an entry to the metrics Collector
//   - Write a structured line to the audit Logger
//   - Publish an event on the event Bus
package observe
