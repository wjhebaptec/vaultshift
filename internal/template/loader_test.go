package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vaultshift/internal/template"
)

func writeTmpl(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("writeTmpl: %v", err)
	}
}

func TestLoadFile_ReadsContent(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "db.tmpl", `{{index . "user"}}`)

	l := template.NewLoader(dir)
	content, err := l.LoadFile("db.tmpl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != `{{index . "user"}}` {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestLoadFile_MissingFile(t *testing.T) {
	l := template.NewLoader(t.TempDir())
	_, err := l.LoadFile("nonexistent.tmpl")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadDir_ReturnsAllTmplFiles(t *testing.T) {
	dir := t.TempDir()
	writeTmpl(t, dir, "alpha.tmpl", "value-alpha")
	writeTmpl(t, dir, "beta.tmpl", "value-beta")
	// non-tmpl file should be ignored
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	l := template.NewLoader(dir)
	templates, err := l.LoadDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
	if templates["alpha"] != "value-alpha" {
		t.Errorf("expected 'value-alpha', got %q", templates["alpha"])
	}
	if templates["beta"] != "value-beta" {
		t.Errorf("expected 'value-beta', got %q", templates["beta"])
	}
}

func TestLoadDir_MissingDir(t *testing.T) {
	l := template.NewLoader("/nonexistent/path")
	_, err := l.LoadDir()
	if err == nil {
		t.Fatal("expected error for missing directory, got nil")
	}
}
