package upload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

func TestAlbumResolverCreation(t *testing.T) {
	resolver := NewAlbumResolver(&flickr.Client{})
	if resolver == nil {
		t.Fatal("expected non-nil resolver")
	}
	if resolver.cache == nil {
		t.Error("expected non-nil cache")
	}
}

func TestAlbumResolverLoad(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Test Album"},"photos":10}]}}`))
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	resolver := NewAlbumResolver(client)
	err := resolver.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resolver.cache) != 1 {
		t.Errorf("expected 1 album in cache, got %d", len(resolver.cache))
	}
}

func TestAlbumResolverResolveOrCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"id":"new-album-123"}}`))
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	resolver := NewAlbumResolver(client)
	id, created, err := resolver.ResolveOrCreate(context.Background(), "New Album", "photo-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "new-album-123" {
		t.Errorf("expected new-album-123, got %s", id)
	}
	if !created {
		t.Error("expected created=true")
	}
}

func TestAlbumResolverResolveOrCreateEmptyTitle(t *testing.T) {
	resolver := NewAlbumResolver(&flickr.Client{})
	_, _, err := resolver.ResolveOrCreate(context.Background(), "", "photo-123")
	if err == nil {
		t.Error("expected error for empty title")
	}
}
