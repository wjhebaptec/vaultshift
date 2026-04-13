package index

import "context"

// Provider is the minimal interface required to build an index from a secret store.
type Provider interface {
	List(ctx context.Context) ([]string, error)
}

// Build populates the index by listing all keys from the given named providers.
// Errors from individual providers are collected and returned as a map; the
// index is still populated for any provider that succeeded.
func Build(ctx context.Context, idx *Index, providers map[string]Provider) map[string]error {
	errs := make(map[string]error)
	for name, p := range providers {
		keys, err := p.List(ctx)
		if err != nil {
			errs[name] = err
			continue
		}
		for _, k := range keys {
			// best-effort: ignore per-key add errors (e.g. empty key)
			_ = idx.Add(name, k)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
