// Package stagger provides time-staggered sequential execution of named tasks.
//
// It is useful when rotating or syncing secrets across many targets and you
// want to spread the load over time rather than hammering all providers
// simultaneously.
//
// Basic usage:
//
//	r := stagger.New(stagger.WithDelay(2 * time.Second))
//	_ = r.Add("aws", func(ctx context.Context) error { return rotateAWS(ctx) })
//	_ = r.Add("gcp", func(ctx context.Context) error { return rotateGCP(ctx) })
//	_ = r.Add("vault", func(ctx context.Context) error { return rotateVault(ctx) })
//
//	results := r.Run(ctx)
//	if stagger.HasFailures(results) {
//		// handle errors
//	}
package stagger
