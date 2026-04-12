// Package propagate implements cross-provider secret propagation for
// vaultshift.
//
// A Propagator holds a set of Rules, each describing a (sourceProvider,
// sourceKey) → (destProvider, destKey) mapping. Calling Propagate or
// PropagateAll reads every source secret and writes it to the corresponding
// destination, enabling secrets to be mirrored or renamed across AWS Secrets
// Manager, GCP Secret Manager, HashiCorp Vault, or any registered provider.
//
// Example:
//
//	prop := propagate.New(registry)
//	prop.AddRule(propagate.Rule{
//		SourceProvider: "aws",
//		SourceKey:      "prod/db/password",
//		DestProvider:   "vault",
//		DestKey:        "secret/db/password",
//	})
//	if err := prop.Propagate(ctx); err != nil {
//		log.Fatal(err)
//	}
package propagate
