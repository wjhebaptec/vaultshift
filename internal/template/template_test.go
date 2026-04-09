package template_test

import (
	"strings"
	"testing"

	"github.com/vaultshift/internal/template"
)

func TestRender_SimpleSubstitution(t *testing.T) {
	r := template.New()
	out, err := r.Render(`Hello, {{index . "name"}}!`, map[string]string{"name": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", out)
	}
}

func TestRender_UpperHelper(t *testing.T) {
	r := template.New()
	out, err := r.Render(`{{upper (index . "env")}}`, map[string]string{"env": "production"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "PRODUCTION" {
		t.Errorf("expected 'PRODUCTION', got %q", out)
	}
}

func TestRender_PrefixHelper(t *testing.T) {
	r := template.New()
	out, err := r.Render(`{{prefix "prod/" (index . "key")}}`, map[string]string{"key": "db-password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "prod/db-password" {
		t.Errorf("expected 'prod/db-password', got %q", out)
	}
}

func TestRender_InvalidTemplate(t *testing.T) {
	r := template.New()
	_, err := r.Render(`{{.Unclosed`, map[string]string{})
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "template parse error") {
		t.Errorf("expected parse error message, got %v", err)
	}
}

func TestRenderAll_RendersAllKeys(t *testing.T) {
	r := template.New()
	templates := map[string]string{
		"db_user": `{{index . "user"}}`,
		"db_pass": `{{upper (index . "pass")}}`,
	}
	data := map[string]string{"user": "admin", "pass": "secret"}

	results, err := r.RenderAll(templates, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["db_user"] != "admin" {
		t.Errorf("expected 'admin', got %q", results["db_user"])
	}
	if results["db_pass"] != "SECRET" {
		t.Errorf("expected 'SECRET', got %q", results["db_pass"])
	}
}

func TestRenderAll_PropagatesError(t *testing.T) {
	r := template.New()
	templates := map[string]string{
		"good": `{{index . "k"}}`,
		"bad":  `{{.Broken`,
	}
	_, err := r.RenderAll(templates, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error from bad template, got nil")
	}
}
