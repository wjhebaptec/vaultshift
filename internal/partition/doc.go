// Package partition provides a secret routing layer that distributes keys
// across multiple cloud secret manager providers based on a user-supplied
// RouterFunc.
//
// A RouterFunc receives a secret key and returns the name of the provider
// that should handle it. This enables use-cases such as:
//
//   - Storing secrets with a "prod/" prefix in AWS and "dev/" prefix in GCP.
//   - Sharding high-volume workloads across several Vault clusters.
//   - Isolating sensitive credentials in a dedicated, higher-security backend.
//
// Example:
//
//	router := func(key string) string {
//		if strings.HasPrefix(key, "prod/") {
//			return "aws"
//		}
//		return "gcp"
//	}
//
//	p, err := partition.New(router, map[string]provider.Provider{
//		"aws": awsProvider,
//		"gcp": gcpProvider,
//	})
package partition
