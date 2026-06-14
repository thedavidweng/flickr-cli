package cli

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func TestCacheHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"cache", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestCacheStatsNoCache(t *testing.T) {
	// Use a temp dir so DefaultCachePath returns a writable location.
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	_, cfg := setupFakeCLI(t)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	app := &AppContext{
		ConfigFile:  cfg,
		Profile:     "default",
		JSON:        true,
		Timeout:     30 * time.Second,
		Retries:     3,
		Concurrency: 4,
		RequestID:   uuid.New().String(),
		StartedAt:   time.Now(),
	}
	cmd.SetContext(WithAppContext(context.Background(), app))

	err := cacheStatsCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "cache.stats" {
		t.Errorf("expected command=cache.stats, got %s", env.Meta.Command)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	counts, ok := data["counts"].(map[string]any)
	if !ok {
		t.Fatalf("expected data.counts to be a map, got %T", data["counts"])
	}
	albumsCount, ok := counts["albums"].(float64)
	if !ok {
		t.Fatal("expected float64 for albums count")
	}
	if albumsCount != 0 {
		t.Errorf("expected 0 albums, got %v", counts["albums"])
	}
	photosCount, ok := counts["photos"].(float64)
	if !ok {
		t.Fatal("expected float64 for photos count")
	}
	if photosCount != 0 {
		t.Errorf("expected 0 photos, got %v", counts["photos"])
	}
	checksumsCount, ok := counts["checksums"].(float64)
	if !ok {
		t.Fatal("expected float64 for checksums count")
	}
	if checksumsCount != 0 {
		t.Errorf("expected 0 checksums, got %v", counts["checksums"])
	}
	jobsCount, ok := counts["jobs"].(float64)
	if !ok {
		t.Fatal("expected float64 for jobs count")
	}
	if jobsCount != 0 {
		t.Errorf("expected 0 jobs, got %v", counts["jobs"])
	}

	// file_bytes should be present and positive (schema creation writes data).
	fileBytes, ok := data["file_bytes"].(float64)
	if !ok {
		t.Fatalf("expected file_bytes to be a number, got %T", data["file_bytes"])
	}
	if fileBytes <= 0 {
		t.Errorf("expected positive file_bytes, got %v", fileBytes)
	}
}

func TestCacheSyncDryRun(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Albums["a1"] = testutil.FakeAlbum{ID: "a1", Title: "Album 1", PhotoCount: 5}
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Photo 1", Owner: "test-user-123"}

	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.Flags().Bool("albums", true, "")
	cmd.Flags().Bool("photos", true, "")
	app := &AppContext{
		ConfigFile:  cfg,
		Profile:     "default",
		JSON:        true,
		DryRun:      true,
		Timeout:     30 * time.Second,
		Retries:     3,
		Concurrency: 4,
		RequestID:   uuid.New().String(),
		StartedAt:   time.Now(),
	}
	cmd.SetContext(WithAppContext(context.Background(), app))

	err := cacheSyncCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "cache.sync" {
		t.Errorf("expected command=cache.sync, got %s", env.Meta.Command)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	// Sync is read-only API + local write; dry-run should not prevent it.
	if _, ok := data["albums_synced"]; !ok {
		t.Error("expected albums_synced in response")
	}
	if _, ok := data["photos_synced"]; !ok {
		t.Error("expected photos_synced in response")
	}

	// Verify the fake server was actually called (sync happened despite dry-run).
	if fake.CountMethod("flickr.photosets.getList") != 1 {
		t.Errorf("expected 1 call to getList, got %d", fake.CountMethod("flickr.photosets.getList"))
	}
	if fake.CountMethod("flickr.people.getPhotos") != 1 {
		t.Errorf("expected 1 call to getPhotos, got %d", fake.CountMethod("flickr.people.getPhotos"))
	}
}
