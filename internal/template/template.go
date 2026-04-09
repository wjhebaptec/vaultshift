// Package template provides secret value templating for vaultshift,
// allowing secret values to be constructed from patterns and references.
package template

import (
	"bytes"
	"fmt"
	"strings"
	gotemplate "text/template"
)

// Renderer renders secret values from Go templates with a provided data context.
type Renderer struct {
	funcs gotemplate.FuncMap
}

// New returns a new Renderer with built-in helper functions.
func New() *Renderer {
	return &Renderer{
		funcs: gotemplate.FuncMap{
			"upper":   strings.ToUpper,
			"lower":   strings.ToLower,
			"trim":    strings.TrimSpace,
			"replace": strings.ReplaceAll,
			"join":    strings.Join,
			"prefix": func(p, s string) string { return p + s },
			"suffix": func(s, sfx string) string { return s + sfx },
		},
	}
}

// Render executes the given template string with the provided data map,
// returning the rendered string or an error.
func (r *Renderer) Render(tmpl string, data map[string]string) (string, error) {
	t, err := gotemplate.New("secret").Funcs(r.funcs).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}

// RenderAll renders a map of templates using the same data context,
// returning a map of rendered values or the first error encountered.
func (r *Renderer) RenderAll(templates map[string]string, data map[string]string) (map[string]string, error) {
	results := make(map[string]string, len(templates))
	for key, tmpl := range templates {
		val, err := r.Render(tmpl, data)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", key, err)
		}
		results[key] = val
	}
	return results, nil
}
