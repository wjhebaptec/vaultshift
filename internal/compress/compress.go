// Package compress provides optional compression for secret values
// before they are stored in or retrieved from a secret manager backend.
package compress

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

// Codec compresses and decompresses secret values.
type Codec struct {
	level int
}

// Option configures a Codec.
type Option func(*Codec)

// WithLevel sets the gzip compression level (e.g. gzip.BestSpeed,
// gzip.BestCompression, gzip.DefaultCompression).
func WithLevel(level int) Option {
	return func(c *Codec) {
		c.level = level
	}
}

// New returns a Codec with the supplied options applied.
func New(opts ...Option) *Codec {
	c := &Codec{level: gzip.DefaultCompression}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Compress gzip-compresses src and returns a base64-encoded string so
// that the result is safe to store as a plain-text secret value.
func (c *Codec) Compress(src string) (string, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return "", fmt.Errorf("compress: create writer: %w", err)
	}
	if _, err = io.WriteString(w, src); err != nil {
		return "", fmt.Errorf("compress: write: %w", err)
	}
	if err = w.Close(); err != nil {
		return "", fmt.Errorf("compress: close writer: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Decompress reverses Compress: it base64-decodes src then gunzips it.
func (c *Codec) Decompress(src string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", fmt.Errorf("compress: base64 decode: %w", err)
	}
	r, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("compress: create reader: %w", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("compress: read: %w", err)
	}
	return string(out), nil
}

// RoundTrip is a convenience helper that compresses then immediately
// decompresses value, useful for verifying codec integrity.
func (c *Codec) RoundTrip(value string) (string, error) {
	compressed, err := c.Compress(value)
	if err != nil {
		return "", err
	}
	return c.Decompress(compressed)
}
