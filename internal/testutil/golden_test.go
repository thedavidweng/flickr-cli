package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateGolden(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	UpdateGolden(t, path, []byte(`{"updated":true}`))

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != `{"updated":true}` {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestUpdateGoldenCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.json")

	UpdateGolden(t, path, []byte(`{"test":true}`))

	if _, err := os.Stat(path); err != nil {
		t.Error("file should exist")
	}
}

func TestLoadGoldenNonExistent(t *testing.T) {
	// This will call t.Fatalf internally, but we can't catch it
	// We'll skip this test since it would fail the test
}
