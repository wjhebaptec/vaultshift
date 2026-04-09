package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// Reporter formats and writes collected metrics to an output sink.
type Reporter struct {
	collector *Collector
	out       io.Writer
}

// NewReporter returns a Reporter backed by the given Collector.
// If out is nil, os.Stdout is used.
func NewReporter(c *Collector, out io.Writer) *Reporter {
	if out == nil {
		out = os.Stdout
	}
	return &Reporter{collector: c, out: out}
}

// PrintTable writes a human-readable summary table.
func (r *Reporter) PrintTable() {
	summary := r.collector.Summary()
	w := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EVENT TYPE\tCOUNT")
	fmt.Fprintln(w, "----------\t-----")
	for _, et := range []EventType{EventRotation, EventSync, EventError} {
		fmt.Fprintf(w, "%s\t%d\n", et, summary[et])
	}
	w.Flush()
}

// PrintJSON writes all entries as a JSON array.
func (r *Reporter) PrintJSON() error {
	entries := r.collector.All()
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
