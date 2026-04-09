// Package audit provides structured, append-only audit logging for vaultshift
// operations. Every secret rotation, sync, deletion, and access event is
// recorded as a newline-delimited JSON entry so that operators can maintain a
// tamper-evident history of changes across all configured secret providers.
//
// Basic usage:
//
//	// Write audit events to a file.
//	f, err := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer f.Close()
//
//	logger := audit.New(f)
//	logger.Log(audit.EventRotate, "aws", "prod/db/password", true, "")
//
// Each event contains a UTC timestamp, the event type, the target provider
// name, the secret key, a success flag, and an optional human-readable message.
package audit
