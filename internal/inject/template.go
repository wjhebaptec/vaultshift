package inject

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
)

// secretRef matches placeholders of the form {{ secret "key" }}.
var secretRef = regexp.MustCompile(`\{\{\s*secret\s+"([^"]+)"\s*\}\}`)

// TemplateInjector resolves secret placeholders embedded inside a text template.
type TemplateInjector struct {
	injector *Injector
}

// NewTemplateInjector creates a TemplateInjector backed by the given Provider.
func NewTemplateInjector(p Provider, opts ...Option) (*TemplateInjector, error) {
	inj, err := New(p, opts...)
	if err != nil {
		return nil, err
	}
	return &TemplateInjector{injector: inj}, nil
}

// Render replaces all {{ secret "key" }} placeholders in src with the
// corresponding secret values fetched from the provider.
func (t *TemplateInjector) Render(ctx context.Context, src string) (string, error) {
	matches := secretRef.FindAllStringSubmatch(src, -1)
	if len(matches) == 0 {
		return src, nil
	}

	keys := make([]string, 0, len(matches))
	seen := map[string]bool{}
	for _, m := range matches {
		if !seen[m[1]] {
			keys = append(keys, m[1])
			seen[m[1]] = true
		}
	}

	resolved, errs := t.injector.InjectMap(ctx, keys)
	if len(errs) > 0 {
		return "", fmt.Errorf("inject: template resolution errors: %v", errs)
	}

	result := secretRef.ReplaceAllFunc([]byte(src), func(match []byte) []byte {
		sub := secretRef.FindSubmatch(match)
		if sub == nil {
			return match
		}
		if v, ok := resolved[string(sub[1])]; ok {
			return []byte(v)
		}
		return match
	})
	return string(bytes.TrimRight(result, "")), nil
}
