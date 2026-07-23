package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB is the cache database.
type DB struct {
	conn    *sql.DB
	profile string
}

// Open opens or creates the cache database.
func Open(path, profile string) (*DB, error) {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating cache dir: %w", err)
		}
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening cache: %w", err)
	}

	if _, err := conn.Exec(Schema); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return &DB{conn: conn, profile: profile}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Stats returns cache statistics.
func (db *DB) Stats() (*CacheStats, error) {
	stats := &CacheStats{}

	var err error
	stats.Counts.Albums, err = db.countAlbums()
	if err != nil {
		return nil, err
	}
	stats.Counts.Photos, err = db.countPhotos()
	if err != nil {
		return nil, err
	}
	stats.Counts.Checksums, err = db.countChecksums()
	if err != nil {
		return nil, err
	}
	stats.Counts.Jobs, err = db.countJobs()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (db *DB) countAlbums() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM albums WHERE profile=?", db.profile).Scan(&count)
	return count, err
}

func (db *DB) countPhotos() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM photos WHERE profile=?", db.profile).Scan(&count)
	return count, err
}

func (db *DB) countChecksums() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM checksums WHERE profile=?", db.profile).Scan(&count)
	return count, err
}

func (db *DB) countJobs() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM jobs WHERE profile=?", db.profile).Scan(&count)
	return count, err
}

// UpsertAlbum inserts or updates an album in the cache.
func (db *DB) UpsertAlbum(id, title, payloadJSON string) error {
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO albums (profile, id, title, payload_json, updated_at) VALUES (?, ?, ?, ?, ?)",
		db.profile, id, title, payloadJSON, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// UpsertPhoto inserts or updates a photo in the cache.
func (db *DB) UpsertPhoto(id, payloadJSON string) error {
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO photos (profile, id, payload_json, updated_at) VALUES (?, ?, ?, ?)",
		db.profile, id, payloadJSON, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// UpsertChecksum inserts or updates a checksum.
func (db *DB) UpsertChecksum(photoID, algorithm, value string) error {
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO checksums (profile, photo_id, algorithm, value, updated_at) VALUES (?, ?, ?, ?, ?)",
		db.profile, photoID, algorithm, value, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// CacheStats contains cache statistics.
type CacheStats struct {
	Path   string `json:"path"`
	Counts struct {
		Albums    int `json:"albums"`
		Photos    int `json:"photos"`
		Checksums int `json:"checksums"`
		Jobs      int `json:"jobs"`
	} `json:"counts"`
}

// Cleanup removes old entries from the cache.
func (db *DB) Cleanup(olderThan time.Duration) (int, error) {
	cutoff := time.Now().UTC().Add(-olderThan).Format(time.RFC3339)

	result, err := db.conn.Exec("DELETE FROM jobs WHERE updated_at < ?", cutoff)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("getting rows affected: %w", err)
	}
	return int(rows), nil
}

// StatFile returns the size of the cache file at the given path.
func StatFile(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
