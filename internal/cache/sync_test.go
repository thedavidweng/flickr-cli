package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

func TestSyncOptions(t *testing.T) {
	opts := SyncOptions{
		Albums: true,
		Photos: false,
		Limit:  100,
	}

	if !opts.Albums {
		t.Error("expected albums=true")
	}
	if opts.Photos {
		t.Error("expected photos=false")
	}
	if opts.Limit != 100 {
		t.Errorf("expected limit 100, got %d", opts.Limit)
	}
}

func TestSyncResult(t *testing.T) {
	result := &SyncResult{
		AlbumsSynced: 10,
		PhotosSynced: 50,
	}

	if result.AlbumsSynced != 10 {
		t.Errorf("expected albums synced 10, got %d", result.AlbumsSynced)
	}
	if result.PhotosSynced != 50 {
		t.Errorf("expected photos synced 50, got %d", result.PhotosSynced)
	}
}

func TestSyncAlbums(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Test Album"},"photos":10}]}}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer db.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	result, err := Sync(context.Background(), db, client, SyncOptions{Albums: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AlbumsSynced != 1 {
		t.Errorf("expected 1 album synced, got %d", result.AlbumsSynced)
	}
}

func TestSyncPhotos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Test Photo"}]}}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer db.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	result, err := Sync(context.Background(), db, client, SyncOptions{Photos: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PhotosSynced != 1 {
		t.Errorf("expected 1 photo synced, got %d", result.PhotosSynced)
	}
}
