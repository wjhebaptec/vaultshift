// Package bridge implements bidirectional secret forwarding between two
// registered providers.
//
// A Bridge copies secrets from a source provider to a destination provider,
// optionally transforming key names or filtering which keys are forwarded.
//
// Example — one-way forward:
//
//	b, err := bridge.New(reg, "aws", "gcp")
//	if err != nil { ... }
//	b.WithFilter(func(k string) bool { return strings.HasPrefix(k, "prod/") })
//	results, err := b.Forward(ctx)
//
// Example — bidirectional sync:
//
//	res, err := bridge.Sync(ctx, reg, "vault", "aws")
//	if err != nil { ... }
//	if res.HasFailures() { ... }
package bridge
