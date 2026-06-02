package upload

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanStableOrder(t *testing.T) {
	dir := t.TempDir()
	files := []string{"c.jpg", "a.jpg", "b.jpg"}
	for _, f := range files {
		os.WriteFile(filepath.Join(dir, f), []byte("test"), 0o644)
	}

	valid, _, err := Scan([]string{dir}, ScanOptions{Recursive: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(valid) != 3 {
		t.Fatalf("expected 3 files, got %d", len(valid))
	}

	// Should be sorted by path
	if filepath.Base(valid[0].Path) != "a.jpg" {
		t.Errorf("expected a.jpg first, got %s", filepath.Base(valid[0].Path))
	}
	if filepath.Base(valid[1].Path) != "b.jpg" {
		t.Errorf("expected b.jpg second, got %s", filepath.Base(valid[1].Path))
	}
	if filepath.Base(valid[2].Path) != "c.jpg" {
		t.Errorf("expected c.jpg third, got %s", filepath.Base(valid[2].Path))
	}
}

func TestScanInvalidExtensions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(dir, "doc.txt"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(dir, "image.png"), []byte("test"), 0o644)

	valid, invalid, err := Scan([]string{dir}, ScanOptions{Recursive: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(valid) != 2 {
		t.Errorf("expected 2 valid files, got %d", len(valid))
	}
	if len(invalid) != 1 {
		t.Errorf("expected 1 invalid file, got %d", len(invalid))
	}
	if len(invalid) > 0 && invalid[0].Ext != "txt" {
		t.Errorf("expected txt extension, got %s", invalid[0].Ext)
	}
}

func TestScanSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")
	os.WriteFile(path, []byte("test"), 0o644)

	valid, _, err := Scan([]string{path}, ScanOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(valid) != 1 {
		t.Fatalf("expected 1 file, got %d", len(valid))
	}
	if valid[0].Size != 4 {
		t.Errorf("expected size 4, got %d", valid[0].Size)
	}
}

func TestScanRecursive(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0o755)
	os.WriteFile(filepath.Join(dir, "a.jpg"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(subdir, "b.jpg"), []byte("test"), 0o644)

	valid, _, err := Scan([]string{dir}, ScanOptions{Recursive: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(valid) != 2 {
		t.Errorf("expected 2 files with recursive, got %d", len(valid))
	}
}

func TestScanNonRecursiveDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	os.Mkdir(subdir, 0o755)
	os.WriteFile(filepath.Join(dir, "a.jpg"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(subdir, "b.jpg"), []byte("test"), 0o644)

	valid, _, err := Scan([]string{dir}, ScanOptions{Recursive: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(valid) != 1 {
		t.Errorf("expected 1 file without recursive, got %d", len(valid))
	}
}

func TestScanCustomExtensions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(dir, "photo.raw"), []byte("test"), 0o644)

	valid, invalid, err := Scan([]string{dir}, ScanOptions{
		AcceptedExt: []string{"raw"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(valid) != 1 {
		t.Errorf("expected 1 valid file, got %d", len(valid))
	}
	if len(invalid) != 1 {
		t.Errorf("expected 1 invalid file, got %d", len(invalid))
	}
}
