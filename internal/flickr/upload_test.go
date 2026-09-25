package flickr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rsp stat="ok"><photoid>12345</photoid></rsp>`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
		HTTP:      server.Client(),
		Endpoints: Endpoints{Upload: server.URL + "/upload"},
	}

	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.jpg"
	if err := os.WriteFile(tmpFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := client.Upload(context.Background(), tmpFile, &UploadOptions{Title: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PhotoID != "12345" {
		t.Errorf("expected photo ID 12345, got %s", result.PhotoID)
	}
}

func TestAddToAlbum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	err := client.AddToAlbum(context.Background(), "album-123", "photo-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadSignatureExcludesPhoto(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rsp stat="ok"><photoid>99999</photoid></rsp>`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:      "test-key",
		APISecret:   "test-secret",
		OAuthToken:  "test-token",
		OAuthSecret: "test-token-secret",
		HTTP:        server.Client(),
		Endpoints:   Endpoints{Upload: server.URL + "/upload"},
	}

	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.jpg"
	if err := os.WriteFile(tmpFile, []byte("fake-image-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := client.Upload(context.Background(), tmpFile, &UploadOptions{
		Title:    "My Photo",
		IsPublic: true,
		Tags:     []string{"nature"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Authorization header starts with "OAuth "
	if !strings.HasPrefix(capturedAuth, "OAuth ") {
		t.Fatalf("expected Authorization header to start with 'OAuth ', got %q", capturedAuth)
	}

	// Parse OAuth params from the Authorization header
	oauthPart := strings.TrimPrefix(capturedAuth, "OAuth ")
	parts := strings.Split(oauthPart, ", ")
	oauthKeys := make(map[string]bool)
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.Trim(kv[0], "\"")
			oauthKeys[key] = true
		}
	}

	// Verify that "photo" is NOT in the signed OAuth parameters
	// The photo file is in the multipart body, not in the signature
	if oauthKeys["photo"] {
		t.Error("expected 'photo' to NOT be in signed OAuth parameters")
	}

	// Verify that other expected params ARE present
	for _, expected := range []string{"oauth_consumer_key", "oauth_signature", "oauth_token"} {
		if !oauthKeys[expected] {
			t.Errorf("expected '%s' to be in OAuth parameters", expected)
		}
	}
}
