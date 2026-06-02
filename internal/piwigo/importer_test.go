package piwigo

import (
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/safety"
)

func TestImporterDryRun(t *testing.T) {
	imp := &Importer{
		Gate: safety.GateInput{DryRun: true},
	}

	opts := ImportOptions{
		DB: DBConfig{DB: "test", User: "root"},
	}

	summary, err := imp.Import(nil, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Planned != 0 {
		t.Errorf("expected planned=0 for dry-run, got %d", summary.Planned)
	}
}

func TestImporterReadOnly(t *testing.T) {
	imp := &Importer{
		Gate: safety.GateInput{ReadOnly: true},
	}

	opts := ImportOptions{
		DB: DBConfig{DB: "test", User: "root"},
	}

	_, err := imp.Import(nil, opts)
	if err == nil {
		t.Error("expected error for read-only mode")
	}
}

func TestImporterInvalidDB(t *testing.T) {
	imp := &Importer{}

	opts := ImportOptions{
		DB: DBConfig{DB: "", User: ""},
	}

	_, err := imp.Import(nil, opts)
	if err == nil {
		t.Error("expected error for invalid DB config")
	}
}
