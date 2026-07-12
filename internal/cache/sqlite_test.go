package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, err := Open(path, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database file not created: %v", err)
	}
}

func TestUpsertAlbum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

	if err := db.UpsertAlbum("123", "My Album", `{"id":"123"}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Counts.Albums != 1 {
		t.Errorf("expected 1 album, got %d", stats.Counts.Albums)
	}
}

func TestUpsertPhoto(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

	if err := db.UpsertPhoto("456", `{"id":"456"}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats, _ := db.Stats()
	if stats.Counts.Photos != 1 {
		t.Errorf("expected 1 photo, got %d", stats.Counts.Photos)
	}
}

func TestUpsertChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

	if err := db.UpsertChecksum("456", "md5", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats, _ := db.Stats()
	if stats.Counts.Checksums != 1 {
		t.Errorf("expected 1 checksum, got %d", stats.Counts.Checksums)
	}
}

func TestCleanup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

	count, err := db.Cleanup(24 * time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 cleaned, got %d", count)
	}
}

func TestCleanupOnClosedDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	_ = db.Close()

	_, err := db.Cleanup(24 * time.Hour)
	if err == nil {
		t.Error("expected error on closed DB, got nil")
	}
}

func TestStatFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(path, []byte("hello"), 0o644)

	size, err := StatFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size != 5 {
		t.Errorf("expected size 5, got %d", size)
	}
}

func TestStatFileNonExistent(t *testing.T) {
	_, err := StatFile("/nonexistent")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestCacheStats(t *testing.T) {
	stats := &CacheStats{
		Path: "/tmp/test.sqlite",
	}
	stats.Counts.Albums = 10
	stats.Counts.Photos = 100
	stats.Counts.Checksums = 50
	stats.Counts.Jobs = 2

	if stats.Counts.Albums != 10 {
		t.Errorf("expected 10 albums, got %d", stats.Counts.Albums)
	}
}

func TestOpenCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.sqlite")

	db, err := Open(path, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := os.Stat(path); err != nil {
		t.Error("database file should exist")
	}
}

func TestDBClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	if err := db.Close(); err != nil {
		t.Fatalf("unexpected error closing: %v", err)
	}
}

func TestDBStats(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Error("expected non-nil stats")
	}
}
