package backup

import (
	"strings"
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

func TestIDDirsPath(t *testing.T) {
	path := IDDirsPath("/dest", "12345", "jpg")

	if len(path) < 10 {
		t.Errorf("path too short: %s", path)
	}
	if !strings.Contains(path, "12345") {
		t.Errorf("path should contain photo ID: %s", path)
	}
}
