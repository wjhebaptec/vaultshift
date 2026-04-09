package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/vaultshift/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	tmp := filepath.Join(t.TempDir(), "vaultshift.yaml")
	if err := os.WriteFile(tmp, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return tmp
}

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
version: "1"
aws:
  region: us-east-1
sync_rules:
  - name: rotate-db-password
    source_key: aws:///prod/db/password
    target_keys:
      - gcp:///projects/myproject/secrets/db-password
    rotate: true
    rotate_every: 24h
`
	path := writeTemp(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("expected version '1', got %q", cfg.Version)
	}
	if cfg.AWS.Region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got %q", cfg.AWS.Region)
	}
	if len(cfg.SyncRules) != 1 {
		t.Fatalf("expected 1 sync rule, got %d", len(cfg.SyncRules))
	}
	if cfg.SyncRules[0].Name != "rotate-db-password" {
		t.Errorf("unexpected rule name: %q", cfg.SyncRules[0].Name)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/vaultshift.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_MissingVersion(t *testing.T) {
	yaml := `
sync_rules:
  - name: test-rule
    source_key: aws:///prod/key
    target_keys:
      - vault:///secret/key
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing version")
	}
}

func TestLoad_MissingTargetKeys(t *testing.T) {
	yaml := `
version: "1"
sync_rules:
  - name: bad-rule
    source_key: aws:///prod/key
    target_keys: []
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty target_keys")
	}
}
