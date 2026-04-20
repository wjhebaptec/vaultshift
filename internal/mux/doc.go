// Package mux implements a routing multiplexer for secret providers.
//
// A Mux accepts a Router function that maps each secret key to a named
// provider. On every operation the key is passed through the router and
// the call is forwarded to the matching registered provider.
//
// Example usage:
//
//	awsProv := awsprovider.New(cfg)
//	gcpProv := gcpprovider.New(cfg)
//
//	m, _ := mux.New(func(key string) string {
//		if strings.HasPrefix(key, "aws/") {
//			return "aws"
//		}
//		return "gcp"
//	})
//	_ = m.Register("aws", awsProv)
//	_ = m.Register("gcp", gcpProv)
//
//	// Automatically routed to the AWS provider.
//	_ = m.Put(ctx, "aws/db-password", "s3cr3t")
package mux
