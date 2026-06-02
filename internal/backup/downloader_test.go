package backup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

func TestDownloaderCreation(t *testing.T) {
	downloader := &Downloader{
		Concurrency: 4,
		Events:      output.EventWriter{},
	}

	if downloader.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", downloader.Concurrency)
	}
}

func TestDownloadOptions(t *testing.T) {
	opts := DownloadOptions{
		Force:    true,
		Size:     "original",
		Metadata: "json",
	}

	if !opts.Force {
		t.Error("expected force=true")
	}
	if opts.Size != "original" {
		t.Errorf("expected original, got %s", opts.Size)
	}
}

func TestDownloadSummary(t *testing.T) {
	summary := &DownloadSummary{
		Total:     10,
		Completed: 8,
		Skipped:   1,
		Failed:    1,
	}

	if summary.Total != 10 {
		t.Errorf("expected total 10, got %d", summary.Total)
	}
	if summary.Completed != 8 {
		t.Errorf("expected completed 8, got %d", summary.Completed)
	}
}

func TestDownloadItem(t *testing.T) {
	item := DownloadItem{
		PhotoID:          "photo-123",
		FilePath:         "/tmp/photo.jpg",
		MetadataPathJSON: "/tmp/photo.json",
		SizeLabel:        "original",
	}

	if item.PhotoID != "photo-123" {
		t.Errorf("expected photo-123, got %s", item.PhotoID)
	}
}

func TestDownload(t *testing.T) {
	// Create a mock server that returns a photo
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","sizes":{"size":[{"label":"Original","width":4000,"height":3000,"source":"` + photoServer.URL + `/photo.jpg","url":"` + photoServer.URL + `/photo.jpg","media":"photo"}]}}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      client,
		Concurrency: 1,
		Events:      output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "photo-123", FilePath: filePath, SizeLabel: "original"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", summary.Completed)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); err != nil {
		t.Error("downloaded file should exist")
	}
}

func TestDownloadSkipExisting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")
	os.WriteFile(filePath, []byte("existing"), 0o644)

	downloader := &Downloader{
		HTTP:        http.DefaultClient,
		Client:      &flickr.Client{},
		Concurrency: 1,
		Events:      output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "photo-123", FilePath: filePath, SizeLabel: "original"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{Force: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed (skipped), got %d", summary.Completed)
	}
}
