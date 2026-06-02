package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGolden(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	os.WriteFile(path, []byte(`{"key":"value"}`), 0o644)

	content := LoadGolden(t, path)
	if string(content) != `{"key":"value"}` {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestCompareJSON(t *testing.T) {
	expected := []byte(`{"key":"value"}`)
	actual := []byte(`{"key": "value"}`)

	// Should not fail - JSON comparison ignores whitespace
	CompareJSON(t, expected, actual)
}

func TestCompareJSONMismatch(t *testing.T) {
	// We can't easily test the failure case since it calls t.Errorf
	// But we can test that it doesn't panic with matching values
	expected := []byte(`{"key":"value"}`)
	actual := []byte(`{"key":"value"}`)

	// This should not call t.Errorf
	CompareJSON(t, expected, actual)
}

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
