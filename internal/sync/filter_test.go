package sync

import "testing"

func TestPrefixFilter(t *testing.T) {
	filter := NewPrefixFilter("prod/")

	tests := []struct {
		key      string
		expected bool
	}{
		{"prod/api-key", true},
		{"prod/db-password", true},
		{"dev/api-key", false},
		{"staging/key", false},
	}

	for _, tt := range tests {
		result := filter(tt.key)
		if result != tt.expected {
			t.Errorf("PrefixFilter(%s) = %v, want %v", tt.key, result, tt.expected)
		}
	}
}

func TestSuffixFilter(t *testing.T) {
	filter := NewSuffixFilter("-password")

	tests := []struct {
		key      string
		expected bool
	}{
		{"db-password", true},
		{"admin-password", true},
		{"api-key", false},
		{"password", false},
	}

	for _, tt := range tests {
		result := filter(tt.key)
		if result != tt.expected {
			t.Errorf("SuffixFilter(%s) = %v, want %v", tt.key, result, tt.expected)
		}
	}
}

func TestRegexFilter(t *testing.T) {
	filter, err := NewRegexFilter(`^prod/.*-key$`)
	if err != nil {
		t.Fatalf("NewRegexFilter failed: %v", err)
	}

	tests := []struct {
		key      string
		expected bool
	}{
		{"prod/api-key", true},
		{"prod/db-key", true},
		{"prod/password", false},
		{"dev/api-key", false},
	}

	for _, tt := range tests {
		result := filter(tt.key)
		if result != tt.expected {
			t.Errorf("RegexFilter(%s) = %v, want %v", tt.key, result, tt.expected)
		}
	}
}

func TestExcludeFilter(t *testing.T) {
	filter := NewExcludeFilter([]string{"secret1", "secret2"})

	if filter("secret1") {
		t.Error("ExcludeFilter should exclude secret1")
	}
	if filter("secret2") {
		t.Error("ExcludeFilter should exclude secret2")
	}
	if !filter("secret3") {
		t.Error("ExcludeFilter should include secret3")
	}
}

func TestIncludeFilter(t *testing.T) {
	filter := NewIncludeFilter([]string{"secret1", "secret2"})

	if !filter("secret1") {
		t.Error("IncludeFilter should include secret1")
	}
	if !filter("secret2") {
		t.Error("IncludeFilter should include secret2")
	}
	if filter("secret3") {
		t.Error("IncludeFilter should exclude secret3")
	}
}

func TestCombineFilters(t *testing.T) {
	prefixFilter := NewPrefixFilter("prod/")
	suffixFilter := NewSuffixFilter("-key")
	combined := CombineFilters(prefixFilter, suffixFilter)

	if !combined("prod/api-key") {
		t.Error("Combined filter should match prod/api-key")
	}
	if combined("prod/password") {
		t.Error("Combined filter should not match prod/password")
	}
	if combined("dev/api-key") {
		t.Error("Combined filter should not match dev/api-key")
	}
}
