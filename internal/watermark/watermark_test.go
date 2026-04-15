package watermark_test

import (
	"strings"
	"testing"

	"github.com/vaultshift/internal/watermark"
)

func TestNew_EmptyPrefix_ReturnsError(t *testing.T) {
	_, err := watermark.New("")
	if err == nil {
		t.Fatal("expected error for empty prefix")
	}
}

func TestStamp_EmbedsMark(t *testing.T) {
	m, _ := watermark.New("prod")
	stamped := m.Stamp("rotation-v1", "mysecretvalue")
	if !strings.HasPrefix(stamped, "mysecretvalue") {
		t.Errorf("expected value prefix, got %q", stamped)
	}
	if !strings.Contains(stamped, "[wm") {
		t.Errorf("expected watermark tag, got %q", stamped)
	}
}

func TestStrip_RecoversCleanlabel(t *testing.T) {
	m, _ := watermark.New("prod")
	stamped := m.Stamp("rotation-v1", "mysecretvalue")
	clean, label, err := m.Strip(stamped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clean != "mysecretvalue" {
		t.Errorf("expected clean value, got %q", clean)
	}
	if label != "rotation-v1" {
		t.Errorf("expected label 'rotation-v1', got %q", label)
	}
}

func TestStrip_NoWatermark_ReturnsError(t *testing.T) {
	m, _ := watermark.New("prod")
	_, _, err := m.Strip("plainvalue")
	if err == nil {
		t.Fatal("expected error for value without watermark")
	}
}

func TestStrip_PrefixMismatch_ReturnsError(t *testing.T) {
	prod, _ := watermark.New("prod")
	staging, _ := watermark.New("staging")
	stamped := prod.Stamp("v1", "secret")
	_, _, err := staging.Strip(stamped)
	if err == nil {
		t.Fatal("expected prefix mismatch error")
	}
}

func TestContains_True(t *testing.T) {
	m, _ := watermark.New("prod")
	stamped := m.Stamp("v1", "value")
	if !m.Contains(stamped) {
		t.Error("expected Contains to return true")
	}
}

func TestContains_False(t *testing.T) {
	m, _ := watermark.New("prod")
	if m.Contains("plainvalue") {
		t.Error("expected Contains to return false")
	}
}

func TestWithSeparator_CustomSep(t *testing.T) {
	m, _ := watermark.New("prod", watermark.WithSeparator("|"))
	stamped := m.Stamp("v2", "data")
	clean, label, err := m.Strip(stamped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clean != "data" || label != "v2" {
		t.Errorf("unexpected clean=%q label=%q", clean, label)
	}
}
