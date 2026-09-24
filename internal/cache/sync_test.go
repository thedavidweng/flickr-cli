package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

func TestSyncAlbums(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Test Album"},"photos":10}]}}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

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
		_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Test Photo"}],"page":1,"pages":1,"per_page":500,"total":1}}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

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

func TestSyncPhotosPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			// Page 1 of 2: return 2 photos
			_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Photo 1"},{"id":"photo-2","title":"Photo 2"}],"page":1,"pages":2,"per_page":2,"total":3}}`))
		} else {
			// Page 2 of 2: return 1 photo
			_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-3","title":"Photo 3"}],"page":2,"pages":2,"per_page":2,"total":3}}`))
		}
	}))
	defer server.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")

	db, _ := Open(path, "default")
	defer func() { _ = db.Close() }()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	result, err := Sync(context.Background(), db, client, SyncOptions{Photos: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PhotosSynced != 3 {
		t.Errorf("expected 3 photos synced across 2 pages, got %d", result.PhotosSynced)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (one per page), got %d", callCount)
	}
}
