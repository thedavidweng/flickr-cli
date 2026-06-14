package flickr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("key", "secret", "token", "token-secret")

	if client.APIKey != "key" {
		t.Errorf("expected key, got %s", client.APIKey)
	}
	if client.APISecret != "secret" {
		t.Errorf("expected secret, got %s", client.APISecret)
	}
	if client.OAuthToken != "token" {
		t.Errorf("expected token, got %s", client.OAuthToken)
	}
	if client.OAuthSecret != "token-secret" {
		t.Errorf("expected token-secret, got %s", client.OAuthSecret)
	}
	if client.HTTP == nil {
		t.Error("expected non-nil HTTP client")
	}
	if client.UserAgent == "" {
		t.Error("expected non-empty user agent")
	}
}

func TestClientSigner(t *testing.T) {
	client := NewClient("key", "secret", "token", "token-secret")
	signer := client.Signer()

	if signer.Creds.ConsumerKey != "key" {
		t.Errorf("expected key, got %s", signer.Creds.ConsumerKey)
	}
	if signer.Creds.Token != "token" {
		t.Errorf("expected token, got %s", signer.Creds.Token)
	}
}

func TestDefaultEndpoints(t *testing.T) {
	ep := DefaultEndpoints()

	if ep.REST == "" {
		t.Error("expected non-empty REST endpoint")
	}
	if ep.Upload == "" {
		t.Error("expected non-empty Upload endpoint")
	}
	if ep.RequestToken == "" {
		t.Error("expected non-empty RequestToken endpoint")
	}
	if ep.Authorize == "" {
		t.Error("expected non-empty Authorize endpoint")
	}
	if ep.AccessToken == "" {
		t.Error("expected non-empty AccessToken endpoint")
	}
}

func TestClientIsAuthenticated(t *testing.T) {
	client := &Client{OAuthToken: "tok", OAuthSecret: "sec"}
	if !client.IsAuthenticated() {
		t.Error("expected authenticated")
	}

	client2 := &Client{OAuthToken: "tok"}
	if client2.IsAuthenticated() {
		t.Error("expected not authenticated without secret")
	}

	client3 := &Client{}
	if client3.IsAuthenticated() {
		t.Error("expected not authenticated without token")
	}
}

func TestAuthorizationURL(t *testing.T) {
	client := NewClient("key", "secret", "", "")
	url := client.AuthorizationURL("test-token", "read")

	if url == "" {
		t.Error("expected non-empty URL")
	}
	if !contains(url, "oauth_token=test-token") {
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
