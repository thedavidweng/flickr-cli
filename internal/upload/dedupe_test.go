package upload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

func TestDeduplicatorCreation(t *testing.T) {
	dedup := &Deduplicator{
		Algorithm: "md5",
	}
	if dedup.Algorithm != "md5" {
		t.Errorf("expected md5, got %s", dedup.Algorithm)
	}
}

func TestCheckByChecksumFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"existing-photo-123"}],"total":1}}`))
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	dedup := &Deduplicator{
		Client:    client,
		Algorithm: "md5",
	}

	photoID, found, err := dedup.CheckByChecksum(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected found=true")
	}
	if photoID != "existing-photo-123" {
		t.Errorf("expected existing-photo-123, got %s", photoID)
	}
}

func TestCheckByChecksumNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[],"total":0}}`))
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	dedup := &Deduplicator{
		Client:    client,
		Algorithm: "md5",
	}

	_, found, err := dedup.CheckByChecksum(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected found=false")
	}
}
