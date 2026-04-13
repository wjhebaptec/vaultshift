package inject_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vaultshift/internal/inject"
)

func TestRender_NoPlaceholders_ReturnsOriginal(t *testing.T) {
	p := &stubProvider{data: map[string]string{}}
	ti, _ := inject.NewTemplateInjector(p)
	out, err := ti.Render(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Fatalf("expected original string, got %q", out)
	}
}

func TestRender_SinglePlaceholder(t *testing.T) {
	p := &stubProvider{data: map[string]string{"db/pass": "hunter2"}}
	ti, _ := inject.NewTemplateInjector(p)
	src := `password={{ secret "db/pass" }}`
	out, err := ti.Render(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "password=hunter2" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRender_MultiplePlaceholders(t *testing.T) {
	p := &stubProvider{data: map[string]string{"u": "admin", "p": "pass"}}
	ti, _ := inject.NewTemplateInjector(p)
	src := `user={{ secret "u" }} pass={{ secret "p" }}`
	out, err := ti.Render(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "user=admin") || !strings.Contains(out, "pass=pass") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRender_DuplicatePlaceholder_ResolvedOnce(t *testing.T) {
	calls := 0
	p := &countingProvider{data: map[string]string{"k": "v"}, calls: &calls}
	ti, _ := inject.NewTemplateInjector(p)
	src := `{{ secret "k" }} and {{ secret "k" }}`
	out, err := ti.Render(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "v and v" {
		t.Fatalf("unexpected output: %q", out)
	}
	if calls != 1 {
		t.Fatalf("expected 1 provider call, got %d", calls)
	}
}

func TestRender_MissingSecret_ReturnsError(t *testing.T) {
	p := &stubProvider{data: map[string]string{}}
	ti, _ := inject.NewTemplateInjector(p)
	_, err := ti.Render(context.Background(), `{{ secret "missing" }}`)
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

// countingProvider wraps stubProvider and counts Get calls.
type countingProvider struct {
	data  map[string]string
	calls *int
}

func (c *countingProvider) Get(ctx context.Context, key string) (string, error) {
	*c.calls++
	s := &stubProvider{data: c.data}
	return s.Get(ctx, key)
}
