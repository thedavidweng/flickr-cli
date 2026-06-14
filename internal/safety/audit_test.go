package safety

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAppendJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	ev := AuditEvent{
		RequestID: "req-123",
		Profile:   "default",
		Command:   "photos.upload",
		Method:    "flickr.upload",
		Resource:  map[string]any{"path": "/tmp/photo.jpg"},
		Result:    "success",
	}

	if err := Append(path, &ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("audit file not created: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening audit file: %v", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected at least one line")
	}

	var parsed AuditEvent
	if err := json.Unmarshal(scanner.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Command != "photos.upload" {
		t.Errorf("expected command=photos.upload, got %s", parsed.Command)
	}
	if parsed.TS == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestAppendMultiple(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	for i := 0; i < 3; i++ {
		if err := Append(path, &AuditEvent{Command: "test"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening audit file: %v", err)
	}
	defer func() { _ = f.Close() }()

	lines := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
	}
	if lines != 3 {
		t.Errorf("expected 3 lines, got %d", lines)
	}
}

func TestAppendCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "audit.jsonl")

	if err := Append(path, &AuditEvent{Command: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Error("audit file should exist")
	}
}
