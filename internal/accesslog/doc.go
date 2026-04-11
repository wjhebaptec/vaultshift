// Package accesslog provides a lightweight access logging layer for
// vaultshift secret providers.
//
// It records every Get, Put, Delete, and List operation performed against a
// provider, capturing the provider name, secret key, operation type, success
// status, latency, and any error message.
//
// Usage:
//
//	l := accesslog.New()
//	wrapped := accesslog.Wrap("aws", myProvider, l)
//
//	// Use wrapped as a normal provider; all calls are logged.
//	v, err := wrapped.Get(ctx, "db/password")
//
//	// Inspect recorded entries.
//	for _, e := range l.Entries() {
//		fmt.Println(e.Provider, e.Operation, e.Success)
//	}
//
//	// Filter by provider or operation.
//	puts := l.Filter("aws", accesslog.OpPut)
package accesslog
