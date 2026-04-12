// Package relay implements rule-based secret forwarding between providers.
//
// A Relay holds a set of Rules, each describing a source provider+key and a
// destination provider+key. Calling Forward applies every rule in order,
// reading from the source and writing to the destination. An optional
// Transform function may be supplied per-rule to mutate the value in transit.
//
// Example:
//
//	rl := relay.New(registry)
//	_ = rl.Register(relay.Rule{
//		SourceProvider: "aws",
//		SourceKey:      "prod/db/password",
//		DestProvider:   "vault",
//		DestKey:        "secret/db/password",
//	})
//	results := rl.Forward(ctx)
//	if relay.HasFailures(results) { ... }
//
// Use NewFiltered with a RuleFilter to selectively forward only a subset of
// registered rules without modifying the relay's rule list.
package relay
