package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeName(t *testing.T) {
	tests := []struct {
		input    string
		fallback string
		expected string
	}{
		{"hello", "unnamed", "hello"},
		{"hello/world", "unnamed", "hello_world"},
		{"hello\\world", "unnamed", "hello_world"},
		{"hello\x00world", "unnamed", "hello_world"},
		{"", "fallback", "fallback"},
		{"  ", "fallback", "fallback"},
		{"...", "fallback", "fallback"},
		{"CON", "unnamed", "CON_"},
		{"PRN", "unnamed", "PRN_"},
		{"normal-file", "unnamed", "normal-file"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SafeName(tt.input, tt.fallback)
			if got != tt.expected {
				t.Errorf("SafeName(%q, %q) = %q, want %q", tt.input, tt.fallback, got, tt.expected)
			}
		})
	}
}

func TestUniqueName(t *testing.T) {
	dir := t.TempDir()

	// No conflict
	name := UniqueName(dir, "photo.jpg", "123")
	if name != "photo.jpg" {
		t.Errorf("expected photo.jpg, got %s", name)
	}

	// Create conflict
	os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("test"), 0o644)
	name = UniqueName(dir, "photo.jpg", "123")
	if name != "photo-123.jpg" {
		t.Errorf("expected photo-123.jpg, got %s", name)
	}
}

func TestIDDirsPath(t *testing.T) {
	path := IDDirsPath("/dest", "12345", "jpg")

	if len(path) < 10 {
		t.Errorf("path too short: %s", path)
	}
	if !contains(path, "12345") {
		t.Errorf("path should contain photo ID: %s", path)
	}
}

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	data := bytes.NewBufferString("hello world")
	if err := AtomicWriteFile(path, data, 0o644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", content)
	}

	// Temp file should not exist
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("temp file should not exist")
	}
}

func TestAtomicWriteFileCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.txt")

	data := bytes.NewBufferString("test")
	if err := AtomicWriteFile(path, data, 0o644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Error("file should exist")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
