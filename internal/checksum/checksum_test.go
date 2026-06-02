package checksum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileHashMD5(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	hash, err := FileHash(path, "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 32 {
		t.Errorf("expected 32 char md5 hash, got %d chars", len(hash))
	}
	// Known MD5 of "hello world"
	if hash != "5eb63bbbe01eeed093cb22bb8f5acdc3" {
		t.Errorf("unexpected hash: %s", hash)
	}
}

func TestFileHashSHA1(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0o644)

	hash, err := FileHash(path, "sha1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 40 {
		t.Errorf("expected 40 char sha1 hash, got %d chars", len(hash))
	}
}

func TestFileHashInvalidAlgorithm(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("test"), 0o644)

	_, err := FileHash(path, "sha256")
	if err == nil {
		t.Error("expected error for unsupported algorithm")
	}
}

func TestFileHashNonExistentFile(t *testing.T) {
	_, err := FileHash("/nonexistent/file", "md5")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidateAlgorithm(t *testing.T) {
	if err := ValidateAlgorithm("md5"); err != nil {
		t.Errorf("md5 should be valid: %v", err)
	}
	if err := ValidateAlgorithm("sha1"); err != nil {
		t.Errorf("sha1 should be valid: %v", err)
	}
	if err := ValidateAlgorithm("sha256"); err == nil {
		t.Error("sha256 should be invalid")
	}
}
