package environ

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
)

// Writer writes rendered environment variables to an io.Writer.
type Writer struct {
	exporter *Exporter
	w        io.Writer
}

// NewWriter creates a Writer that uses exp to render secrets and writes to w.
// If w is nil, os.Stdout is used.
func NewWriter(exp *Exporter, w io.Writer) *Writer {
	if w == nil {
		w = os.Stdout
	}
	return &Writer{exporter: exp, w: w}
}

// Write renders all secrets from providerName and writes KEY=VALUE lines to
// the underlying writer, one per line, sorted alphabetically.
func (wr *Writer) Write(ctx context.Context, providerName string) error {
	m, err := wr.exporter.Render(ctx, providerName)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := fmt.Fprintf(wr.w, "%s=%s\n", k, m[k]); err != nil {
			return fmt.Errorf("environ: write: %w", err)
		}
	}
	return nil
}
