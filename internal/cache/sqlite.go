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
func Open(path string, profile string) (*DB, error) {
	// Create directory if needed
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
		conn.Close()
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
	stats.Counts.Albums, err = db.count("albums")
	if err != nil {
		return nil, err
	}
	stats.Counts.Photos, err = db.count("photos")
	if err != nil {
		return nil, err
	}
	stats.Counts.Checksums, err = db.count("checksums")
	if err != nil {
		return nil, err
	}
	stats.Counts.Jobs, err = db.count("jobs")
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (db *DB) count(table string) (int, error) {
	var count int
	err := db.conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE profile=?", table), db.profile).Scan(&count)
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

	rows, _ := result.RowsAffected()
	return int(rows), nil
}

// StatFile returns the size of the cache file.
func StatFile(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
