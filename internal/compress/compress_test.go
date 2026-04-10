package compress_test

import (
	"compress/gzip"
	"strings"
	"testing"

	"github.com/vaultshift/internal/compress"
)

func TestCompress_ProducesNonEmptyOutput(t *testing.T) {
	c := compress.New()
	out, err := c.Compress("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty compressed output")
	}
	if out == "hello world" {
		t.Fatal("compressed output should differ from input")
	}
}

func TestDecompress_RestoresOriginal(t *testing.T) {
	c := compress.New()
	original := "super-secret-value-1234"
	enc, err := c.Compress(original)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	got, err := c.Decompress(enc)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if got != original {
		t.Errorf("got %q, want %q", got, original)
	}
}

func TestDecompress_InvalidBase64(t *testing.T) {
	c := compress.New()
	_, err := c.Decompress("!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecompress_ValidBase64ButNotGzip(t *testing.T) {
	c := compress.New()
	// base64 of plain text, not gzip data
	_, err := c.Decompress("aGVsbG8=") // "hello"
	if err == nil {
		t.Fatal("expected error for non-gzip payload")
	}
}

func TestRoundTrip_LargeValue(t *testing.T) {
	c := compress.New()
	large := strings.Repeat("secret-data-", 500)
	got, err := c.RoundTrip(large)
	if err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if got != large {
		t.Errorf("round-trip value mismatch (len got=%d want=%d)", len(got), len(large))
	}
}

func TestWithLevel_BestSpeed(t *testing.T) {
	c := compress.New(compress.WithLevel(gzip.BestSpeed))
	original := "speed-test-value"
	enc, err := c.Compress(original)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	got, err := c.Decompress(enc)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if got != original {
		t.Errorf("got %q, want %q", got, original)
	}
}

func TestCompress_EmptyString(t *testing.T) {
	c := compress.New()
	enc, err := c.Compress("")
	if err != nil {
		t.Fatalf("compress empty: %v", err)
	}
	got, err := c.Decompress(enc)
	if err != nil {
		t.Fatalf("decompress empty: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
