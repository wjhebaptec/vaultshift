package export_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/vaultshift/internal/export"
)

type mockProvider struct {
	secrets map[string]string
	listErr error
	getErr  error
}

func (m *mockProvider) GetSecret(_ context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	v, ok := m.secrets[key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (m *mockProvider) ListSecrets(_ context.Context) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	keys := make([]string, 0, len(m.secrets))
	for k := range m.secrets {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestExport_JSON(t *testing.T) {
	prov := &mockProvider{secrets: map[string]string{"DB_PASS": "secret", "API_KEY": "key123"}}
	var buf bytes.Buffer
	ex := export.New(prov, export.FormatJSON, &buf)
	if err := ex.Export(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["DB_PASS"] != "secret" || got["API_KEY"] != "key123" {
		t.Errorf("unexpected values: %v", got)
	}
}

func TestExport_Env(t *testing.T) {
	prov := &mockProvider{secrets: map[string]string{"TOKEN": "abc", "HOST": "localhost"}}
	var buf bytes.Buffer
	ex := export.New(prov, export.FormatEnv, &buf)
	if err := ex.Export(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "TOKEN=abc") || !strings.Contains(out, "HOST=localhost") {
		t.Errorf("unexpected env output: %s", out)
	}
}

func TestExport_ListError(t *testing.T) {
	prov := &mockProvider{listErr: errors.New("list failed")}
	ex := export.New(prov, export.FormatJSON, &bytes.Buffer{})
	if err := ex.Export(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExport_GetError(t *testing.T) {
	prov := &mockProvider{secrets: map[string]string{"KEY": "val"}, getErr: errors.New("get failed")}
	ex := export.New(prov, export.FormatJSON, &bytes.Buffer{})
	if err := ex.Export(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	prov := &mockProvider{secrets: map[string]string{"K": "v"}}
	ex := export.New(prov, export.Format("xml"), &bytes.Buffer{})
	if err := ex.Export(context.Background()); err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	prov := &mockProvider{secrets: map[string]string{}}
	ex := export.New(prov, export.FormatJSON, nil)
	if ex == nil {
		t.Fatal("expected non-nil exporter")
	}
}
