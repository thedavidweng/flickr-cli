package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// LoadGolden reads a golden file and returns its content.
func LoadGolden(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	return b
}

// CompareJSON compares two JSON byte slices by unmarshaling and re-marshaling
// to normalize formatting differences.
func CompareJSON(t *testing.T, expected, actual []byte) {
	t.Helper()

	var expObj, actObj any

	if err := json.Unmarshal(expected, &expObj); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(actual, &actObj); err != nil {
		t.Fatalf("failed to unmarshal actual JSON: %v\n%s", err, actual)
	}

	expNorm, _ := json.MarshalIndent(expObj, "", "  ")
	actNorm, _ := json.MarshalIndent(actObj, "", "  ")

	if string(expNorm) != string(actNorm) {
		t.Errorf("JSON mismatch:\nExpected:\n%s\nActual:\n%s", expNorm, actNorm)
	}
}

// UpdateGolden writes the actual content to the golden file (for updating tests).
func UpdateGolden(t *testing.T, path string, content []byte) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create golden dir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write golden file %s: %v", path, err)
	}
}
