// Package notify implements a lightweight event notification system for vaultshift.
//
// It allows callers to register one or more Handler functions that are invoked
// whenever a significant operation occurs (rotation, sync, drift detection, failure).
//
// Example usage:
//
//	var buf strings.Builder
//	n := notify.New(notify.LogHandler(func(s string) { buf.WriteString(s + "\n") }))
//	n.Send(notify.Event{
//		Type:     notify.EventRotated,
//		Secret:   "db/password",
//		Provider: "aws",
//		Message:  "rotated successfully",
//	})
package notify
