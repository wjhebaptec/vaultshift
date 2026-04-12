package prefix_test

import (
	"testing"

	"github.com/vaultshift/internal/prefix"
)

func TestNew_EmptyNamespace_ReturnsError(t *testing.T) {
	_, err := prefix.New("")
	if err == nil {
		t.Fatal("expected error for empty namespace")
	}
}

func TestWrap_PrependsPrefix(t *testing.T) {
	p, _ := prefix.New("prod")
	got := p.Wrap("db/password")
	want := "prod/db/password"
	if got != want {
		t.Errorf("Wrap = %q; want %q", got, want)
	}
}

func TestWrap_EmptyKey_ReturnsNamespace(t *testing.T) {
	p, _ := prefix.New("prod")
	if got := p.Wrap(""); got != "prod" {
		t.Errorf("Wrap(\"\") = %q; want \"prod\"", got)
	}
}

func TestWrap_CustomSeparator(t *testing.T) {
	p, _ := prefix.New("prod", prefix.WithSeparator(":"))
	got := p.Wrap("api_key")
	if got != "prod:api_key" {
		t.Errorf("Wrap = %q; want \"prod:api_key\"", got)
	}
}

func TestUnwrap_StripsPrefixCorrectly(t *testing.T) {
	p, _ := prefix.New("prod")
	stripped, ok := p.Unwrap("prod/api_key")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if stripped != "api_key" {
		t.Errorf("Unwrap = %q; want \"api_key\"", stripped)
	}
}

func TestUnwrap_NoPrefix_ReturnsFalse(t *testing.T) {
	p, _ := prefix.New("prod")
	key, ok := p.Unwrap("staging/api_key")
	if ok {
		t.Fatal("expected ok=false for non-matching prefix")
	}
	if key != "staging/api_key" {
		t.Errorf("Unwrap returned modified key %q", key)
	}
}

func TestUnwrap_ExactNamespace_ReturnsEmpty(t *testing.T) {
	p, _ := prefix.New("prod")
	stripped, ok := p.Unwrap("prod")
	if !ok {
		t.Fatal("expected ok=true for exact namespace key")
	}
	if stripped != "" {
		t.Errorf("expected empty string, got %q", stripped)
	}
}

func TestWrapAll_WrapsAllKeys(t *testing.T) {
	p, _ := prefix.New("env")
	input := map[string]string{"key1": "v1", "key2": "v2"}
	out := p.WrapAll(input)
	for _, k := range []string{"env/key1", "env/key2"} {
		if _, ok := out[k]; !ok {
			t.Errorf("expected key %q in wrapped map", k)
		}
	}
}

func TestUnwrapAll_OmitsNonMatchingKeys(t *testing.T) {
	p, _ := prefix.New("env")
	input := map[string]string{
		"env/key1":     "v1",
		"staging/key2": "v2",
	}
	out := p.UnwrapAll(input)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out["key1"] != "v1" {
		t.Errorf("expected out[key1]=v1, got %q", out["key1"])
	}
}

func TestNamespace_ReturnsConfiguredPrefix(t *testing.T) {
	p, _ := prefix.New("myteam")
	if p.Namespace() != "myteam" {
		t.Errorf("Namespace() = %q; want \"myteam\"", p.Namespace())
	}
}
