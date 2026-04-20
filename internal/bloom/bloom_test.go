package bloom

import (
	"testing"
)

func TestNew_DefaultConfig(t *testing.T) {
	f, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
}

func TestNew_InvalidSize_ReturnsError(t *testing.T) {
	_, err := New(WithSize(0))
	if err == nil {
		t.Fatal("expected error for zero size")
	}
}

func TestNew_InvalidHashFunctions_ReturnsError(t *testing.T) {
	_, err := New(WithHashFunctions(0))
	if err == nil {
		t.Fatal("expected error for zero hash functions")
	}
}

func TestAdd_AndMayContain_True(t *testing.T) {
	f, _ := New()
	f.Add("my/secret/key")
	if !f.MayContain("my/secret/key") {
		t.Error("expected MayContain to return true for added key")
	}
}

func TestMayContain_UnaddedKey_LikelyFalse(t *testing.T) {
	f, _ := New(WithSize(4096), WithHashFunctions(4))
	f.Add("present")
	// A key never added should return false (no false positive for this case)
	if f.MayContain("definitely-not-present-xyzzy-12345") {
		t.Log("false positive detected — acceptable but worth noting")
	}
}

func TestMayContain_EmptyFilter_ReturnsFalse(t *testing.T) {
	f, _ := New()
	if f.MayContain("anything") {
		t.Error("empty filter should not report membership")
	}
}

func TestReset_ClearsBits(t *testing.T) {
	f, _ := New()
	f.Add("key1")
	f.Add("key2")
	f.Reset()
	if f.MayContain("key1") {
		t.Error("expected key1 to be absent after reset")
	}
	if f.MayContain("key2") {
		t.Error("expected key2 to be absent after reset")
	}
}

func TestAdd_MultipleKeys(t *testing.T) {
	f, _ := New()
	keys := []string{"alpha", "beta", "gamma", "delta"}
	for _, k := range keys {
		f.Add(k)
	}
	for _, k := range keys {
		if !f.MayContain(k) {
			t.Errorf("expected MayContain(%q) = true", k)
		}
	}
}

func TestAdd_CustomOptions(t *testing.T) {
	f, err := New(WithSize(512), WithHashFunctions(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f.Add("custom")
	if !f.MayContain("custom") {
		t.Error("expected MayContain to return true")
	}
}
