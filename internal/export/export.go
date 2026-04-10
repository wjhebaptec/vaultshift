// Package export provides functionality for exporting secrets to various output formats.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Format represents an output format for exported secrets.
type Format string

const (
	FormatJSON Format = "json"
	FormatEnv  Format = "env"
	FormatDotEnv Format = "dotenv"
)

// Provider is the interface for reading secrets.
type Provider interface {
	GetSecret(ctx context.Context, key string) (string, error)
	ListSecrets(ctx context.Context) ([]string, error)
}

// Exporter writes secrets from a provider to a writer in a given format.
type Exporter struct {
	provider Provider
	format   Format
	out      io.Writer
}

// New creates a new Exporter.
func New(provider Provider, format Format, out io.Writer) *Exporter {
	if out == nil {
		out = os.Stdout
	}
	return &Exporter{provider: provider, format: format, out: out}
}

// Export writes all secrets from the provider to the configured writer.
func (e *Exporter) Export(ctx context.Context) error {
	keys, err := e.provider.ListSecrets(ctx)
	if err != nil {
		return fmt.Errorf("export: list secrets: %w", err)
	}
	sort.Strings(keys)

	secrets := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := e.provider.GetSecret(ctx, k)
		if err != nil {
			return fmt.Errorf("export: get secret %q: %w", k, err)
		}
		secrets[k] = v
	}

	switch e.format {
	case FormatJSON:
		return e.writeJSON(secrets)
	case FormatEnv, FormatDotEnv:
		return e.writeEnv(secrets)
	default:
		return fmt.Errorf("export: unsupported format %q", e.format)
	}
}

func (e *Exporter) writeJSON(secrets map[string]string) error {
	enc := json.NewEncoder(e.out)
	enc.SetIndent("", "  ")
	return enc.Encode(secrets)
}

func (e *Exporter) writeEnv(secrets map[string]string) error {
	var sb strings.Builder
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, secrets[k]))
	}
	_, err := fmt.Fprint(e.out, sb.String())
	return err
}
