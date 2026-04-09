package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Provider represents a supported secret manager backend.
type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderGCP   Provider = "gcp"
	ProviderVault Provider = "vault"
)

// AWSConfig holds AWS Secrets Manager configuration.
type AWSConfig struct {
	Region  string `yaml:"region"`
	Profile string `yaml:"profile,omitempty"`
}

// GCPConfig holds GCP Secret Manager configuration.
type GCPConfig struct {
	Project         string `yaml:"project"`
	CredentialsFile string `yaml:"credentials_file,omitempty"`
}

// VaultConfig holds HashiCorp Vault configuration.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token,omitempty"`
	MountPath string `yaml:"mount_path"`
}

// SyncRule defines a secret rotation/sync rule between providers.
type SyncRule struct {
	Name        string   `yaml:"name"`
	SourceKey   string   `yaml:"source_key"`
	TargetKeys  []string `yaml:"target_keys"`
	Rotate      bool     `yaml:"rotate"`
	RotateEvery string   `yaml:"rotate_every,omitempty"`
}

// Config is the top-level vaultshift configuration.
type Config struct {
	Version   string      `yaml:"version"`
	AWS       AWSConfig   `yaml:"aws,omitempty"`
	GCP       GCPConfig   `yaml:"gcp,omitempty"`
	Vault     VaultConfig `yaml:"vault,omitempty"`
	SyncRules []SyncRule  `yaml:"sync_rules"`
}

// Load reads and parses a vaultshift config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate performs basic sanity checks on the loaded configuration.
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("config version is required")
	}
	for i, rule := range c.SyncRules {
		if rule.Name == "" {
			return fmt.Errorf("sync_rules[%d]: name is required", i)
		}
		if rule.SourceKey == "" {
			return fmt.Errorf("sync_rules[%d] %q: source_key is required", i, rule.Name)
		}
		if len(rule.TargetKeys) == 0 {
			return fmt.Errorf("sync_rules[%d] %q: at least one target_key is required", i, rule.Name)
		}
	}
	return nil
}
