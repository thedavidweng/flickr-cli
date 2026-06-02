package output

import (
	"bytes"
	"testing"
)

func TestTableWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	tw := NewTableWriter(buf)

	tw.Header("ID", "Name", "Count")
	tw.Row("1", "Test", "10")
	tw.Row("2", "Another", "20")

	if err := tw.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}
