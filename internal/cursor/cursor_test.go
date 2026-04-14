package cursor_test

import (
	"strings"
	"testing"

	"github.com/vaultshift/internal/cursor"
)

func TestEncode_Decode_RoundTrip(t *testing.T) {
	m := cursor.New(cursor.WithPageSize(25))
	c := cursor.Cursor{Provider: "aws", Offset: 0, Prefix: "prod/", PageSize: 25}

	token, err := m.Encode(c)
	if err != nil {
		t.Fatalf("Encode: unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	got, err := m.Decode(token)
	if err != nil {
		t.Fatalf("Decode: unexpected error: %v", err)
	}
	if got.Provider != c.Provider || got.Offset != c.Offset || got.Prefix != c.Prefix {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, c)
	}
}

func TestEncode_EmptyProvider_ReturnsError(t *testing.T) {
	m := cursor.New()
	_, err := m.Encode(cursor.Cursor{Provider: ""})
	if err == nil {
		t.Fatal("expected error for empty provider")
	}
}

func TestEncode_DefaultPageSize_Applied(t *testing.T) {
	m := cursor.New(cursor.WithPageSize(10))
	c := cursor.Cursor{Provider: "gcp", PageSize: 0}

	token, err := m.Encode(c)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := m.Decode(token)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", got.PageSize)
	}
}

func TestDecode_EmptyToken_ReturnsError(t *testing.T) {
	m := cursor.New()
	_, err := m.Decode("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestDecode_InvalidBase64_ReturnsError(t *testing.T) {
	m := cursor.New()
	_, err := m.Decode("!!!notbase64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecode_ValidBase64ButNoProvider_ReturnsError(t *testing.T) {
	m := cursor.New()
	// base64 of `{"offset":0}` — missing provider
	import64 := "eyJvZmZzZXQiOjB9"
	_, err := m.Decode(import64)
	if err == nil {
		t.Fatal("expected error when provider is missing")
	}
}

func TestNext_AdvancesOffset(t *testing.T) {
	m := cursor.New(cursor.WithPageSize(20))
	c := cursor.Cursor{Provider: "vault", Offset: 0, PageSize: 20}

	next := m.Next(c)
	if next.Offset != 20 {
		t.Errorf("expected offset 20, got %d", next.Offset)
	}

	next2 := m.Next(next)
	if next2.Offset != 40 {
		t.Errorf("expected offset 40, got %d", next2.Offset)
	}
}

func TestNext_UsesDefaultPageSizeWhenZero(t *testing.T) {
	m := cursor.New(cursor.WithPageSize(15))
	c := cursor.Cursor{Provider: "aws", Offset: 0, PageSize: 0}
	next := m.Next(c)
	if next.Offset != 15 {
		t.Errorf("expected offset 15, got %d", next.Offset)
	}
}

func TestEncode_ProducesURLSafeToken(t *testing.T) {
	m := cursor.New()
	c := cursor.Cursor{Provider: "gcp", Offset: 100, PageSize: 50}
	token, err := m.Encode(c)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if strings.ContainsAny(token, "+/=") {
		t.Errorf("token contains non-URL-safe characters: %s", token)
	}
}
