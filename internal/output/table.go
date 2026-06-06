package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// TableWriter writes tab-aligned tables to an io.Writer.
type TableWriter struct {
	w *tabwriter.Writer
}

// NewTableWriter creates a new TableWriter.
func NewTableWriter(out io.Writer) *TableWriter {
	return &TableWriter{
		w: tabwriter.NewWriter(out, 0, 4, 2, ' ', 0),
	}
}

// Header writes the header row.
func (t *TableWriter) Header(cols ...string) {
	_, _ = fmt.Fprintln(t.w, strings.Join(cols, "\t"))
}

// Row writes a data row.
func (t *TableWriter) Row(cols ...string) {
	_, _ = fmt.Fprintln(t.w, strings.Join(cols, "\t"))
}

// Flush flushes the tabwriter.
func (t *TableWriter) Flush() error {
	return t.w.Flush()
}
