package flickr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthorizationURL(t *testing.T) {
	client := NewClient("key", "secret", "", "")
	url := client.AuthorizationURL("test-token", "read")

	if url == "" {
		t.Error("expected non-empty URL")
	}
	if !strings.Contains(url, "oauth_token=test-token") {
		t.Error("URL should contain oauth_token")
	}
}

func TestRequestToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte("oauth_token=req-token&oauth_token_secret=req-secret&oauth_callback_confirmed=true"))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "key",
		APISecret: "secret",
		HTTP:      server.Client(),
		Endpoints: Endpoints{RequestToken: server.URL},
	}

	resp, err := client.RequestToken(context.Background(), "http://localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "req-token" {
		t.Errorf("expected req-token, got %s", resp.Token)
	}
}

func TestAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte("oauth_token=access-token&oauth_token_secret=access-secret&user_nsid=user123&username=testuser"))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "key",
		APISecret: "secret",
		HTTP:      server.Client(),
		Endpoints: Endpoints{AccessToken: server.URL},
	}

	resp, err := client.AccessToken(context.Background(), "req-token", "req-secret", "verifier")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "access-token" {
		t.Errorf("expected access-token, got %s", resp.Token)
	}
	if resp.UserNSID != "user123" {
		t.Errorf("expected user123, got %s", resp.UserNSID)
	}
}

// contains and containsSubstr are defined in urls.go
