// Package pipeline implements a composable, sequential execution engine
// for chaining secret-management operations in vaultshift.
//
// A Pipeline is an ordered list of Steps. Each Step receives a mutable
// Payload that carries the secret key, provider name, value, and arbitrary
// metadata. Steps are executed in registration order; the pipeline halts
// immediately if any step returns an error or the context is cancelled.
//
// # Basic usage
//
//	p := pipeline.New()
//	p.Add(pipeline.Step{Name: "validate", Run: myValidator})
//	p.Add(pipeline.Step{Name: "rotate",   Run: myRotator})
//	results, err := p.Execute(ctx, payload)
//
// # Builder
//
// The Builder type provides a fluent API for constructing common pipelines
// without manually creating Step structs:
//
//	p := pipeline.NewBuilder().
//		WithValidation("non-empty", func(v string) bool { return v != "" }).
//		WithTransform("upper", func(v string) (string, error) { return strings.ToUpper(v), nil }).
//		WithMetaTag("env", "prod").
//		Build()
package pipeline
