// Package trace provides lightweight operation tracing for vaultshift.
//
// A Tracer records Spans — each representing a discrete secret operation
// (rotate, sync, get, put, delete) against a named provider and key.
//
// Trace IDs are propagated through context so that multiple spans
// originating from a single high-level command share the same trace ID,
// enabling end-to-end correlation in audit logs or external observability
// systems.
//
// Example usage:
//
//	tr := trace.New()
//	ctx, span := tr.Start(ctx, "rotate", "aws", "prod/db/password")
//	// ... perform operation ...
//	tr.Finish(span, err)
//
//	for _, s := range tr.Spans() {
//		fmt.Printf("%s %s %s\n", s.TraceID, s.Operation, s.Key)
//	}
package trace
