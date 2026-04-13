// Package compare provides cross-provider secret comparison for vaultshift.
//
// Use compare.New to create a Comparer backed by a provider.Registry, then
// call Compare or CompareAll to inspect whether a secret key holds the same
// value across every named provider.
//
// Example:
//
//	c, _ := compare.New(registry)
//	res, _ := c.Compare(ctx, "db/password", []string{"aws", "gcp"})
//	if !res.Match {
//		fmt.Println("drift detected:", res.Values)
//	}
package compare
