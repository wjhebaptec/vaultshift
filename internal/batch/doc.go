// Package batch provides bounded, concurrent batch processing for secret
// operations in vaultshift.
//
// A Processor splits a slice of Items into fixed-size chunks and executes a
// user-supplied ProcessFunc against each item, optionally using multiple
// goroutines within each chunk.
//
// Basic usage:
//
//	p := batch.New(
//		batch.WithSize(25),    // up to 25 items per chunk
//		batch.WithWorkers(4),  // 4 concurrent workers per chunk
//	)
//
//	results := p.Run(ctx, items, func(ctx context.Context, item batch.Item) error {
//		return provider.Put(ctx, item.Key, item.Value)
//	})
//
// Each Result in the returned slice corresponds positionally to the input
// Item slice, making it straightforward to correlate successes and failures.
package batch
