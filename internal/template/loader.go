package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader reads template definitions from the filesystem.
type Loader struct {
	baseDir string
}

// NewLoader returns a Loader rooted at baseDir.
func NewLoader(baseDir string) *Loader {
	return &Loader{baseDir: baseDir}
}

// LoadFile reads a single template file and returns its contents as a string.
func (l *Loader) LoadFile(name string) (string, error) {
	path := filepath.Join(l.baseDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("template loader: %w", err)
	}
	return string(data), nil
}

// LoadDir reads all ".tmpl" files from the base directory and returns a map
// of filename-without-extension to template content.
func (l *Loader) LoadDir() (map[string]string, error) {
	entries, err := os.ReadDir(l.baseDir)
	if err != nil {
		return nil, fmt.Errorf("template loader: %w", err)
	}

	templates := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}
		content, err := l.LoadFile(entry.Name())
		if err != nil {
			return nil, err
		}
		key := strings.TrimSuffix(entry.Name(), ".tmpl")
		templates[key] = content
	}
	return templates, nil
}
