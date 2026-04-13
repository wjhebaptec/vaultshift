package pin_test

import (
	"testing"

	"github.com/vaultshift/internal/pin"
)

func TestPin_AndGet(t *testing.T) {
	p := pin.New()
	if err := p.Pin("aws", "db/pass", "v3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, err := p.Get("aws", "db/pass")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if e.Version != "v3" {
		t.Errorf("want version v3, got %s", e.Version)
	}
	if e.PinnedAt.IsZero() {
		t.Error("PinnedAt should be set")
	}
}

func TestGet_NotPinned(t *testing.T) {
	p := pin.New()
	_, err := p.Get("aws", "missing")
	if err != pin.ErrNotPinned {
		t.Errorf("want ErrNotPinned, got %v", err)
	}
}

func TestPin_EmptyProvider_ReturnsError(t *testing.T) {
	p := pin.New()
	if err := p.Pin("", "key", "v1"); err == nil {
		t.Error("expected error for empty provider")
	}
}

func TestPin_EmptyKey_ReturnsError(t *testing.T) {
	p := pin.New()
	if err := p.Pin("aws", "", "v1"); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestPin_EmptyVersion_ReturnsError(t *testing.T) {
	p := pin.New()
	if err := p.Pin("aws", "key", ""); err == nil {
		t.Error("expected error for empty version")
	}
}

func TestUnpin_RemovesEntry(t *testing.T) {
	p := pin.New()
	_ = p.Pin("gcp", "token", "v1")
	p.Unpin("gcp", "token")
	if p.IsPinned("gcp", "token") {
		t.Error("expected key to be unpinned")
	}
}

func TestIsPinned_False(t *testing.T) {
	p := pin.New()
	if p.IsPinned("vault", "secret") {
		t.Error("expected false for unknown key")
	}
}

func TestAll_ReturnsPinnedEntries(t *testing.T) {
	p := pin.New()
	_ = p.Pin("aws", "key1", "v1")
	_ = p.Pin("aws", "key2", "v2")
	entries := p.All()
	if len(entries) != 2 {
		t.Errorf("want 2 entries, got %d", len(entries))
	}
}

func TestPin_SeparateProviders_AreIndependent(t *testing.T) {
	p := pin.New()
	_ = p.Pin("aws", "shared", "v1")
	_ = p.Pin("gcp", "shared", "v9")

	aws, _ := p.Get("aws", "shared")
	gcp, _ := p.Get("gcp", "shared")

	if aws.Version == gcp.Version {
		t.Error("providers should be independent")
	}
}
