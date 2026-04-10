// Package compress provides transparent gzip + base64 compression for
// secret values managed by vaultshift.
//
// Some secret manager backends impose size limits on stored values. The
// Codec in this package can be used to compress large secrets (e.g.
// TLS certificate bundles, JSON blobs) before they are written, and to
// decompress them transparently on read.
//
// Usage:
//
//	c := compress.New(compress.WithLevel(gzip.BestCompression))
//
//	encoded, err := c.Compress(rawSecret)
//	// store encoded in the backend …
//
//	restored, err := c.Decompress(encoded)
//	// use restored as the original secret value …
//
// The compressed payload is base64-encoded so it can be stored safely
// wherever plain-text strings are expected.
package compress
