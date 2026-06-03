package flickr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCallRawSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","result":"success"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	result, err := client.CallRaw(context.Background(), "flickr.test.echo", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCallRawFailResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"fail","code":1,"message":"Not found"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	_, err := client.CallRaw(context.Background(), "flickr.test.fail", nil)
	if err == nil {
		t.Error("expected error for fail response")
	}
}

func TestCallRawHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	_, err := client.CallRaw(context.Background(), "flickr.test.error", nil)
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestCallDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","name":"test"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	var result map[string]string
	err := client.Call(context.Background(), "flickr.test.echo", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %s", result["name"])
	}
}

func TestRESTAddsDefaultParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("method") == "" {
			t.Error("missing method param")
		}
		if q.Get("api_key") == "" {
			t.Error("missing api_key param")
		}
		if q.Get("format") != "json" {
			t.Error("missing format=json")
		}
		if q.Get("nojsoncallback") != "1" {
			t.Error("missing nojsoncallback=1")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	client.CallRaw(context.Background(), "flickr.test.echo", nil)
}

func TestTestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","user":{"id":"test-user-123","username":{"_content":"testuser"}}}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	result, err := client.TestLogin(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.UserNSID != "test-user-123" {
		t.Errorf("expected UserNSID=test-user-123, got %s", result.UserNSID)
	}
	if result.Username != "testuser" {
		t.Errorf("expected Username=testuser, got %s", result.Username)
	}
}

func TestTestEcho(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	err := client.TestEcho(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryHTTP503(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service Unavailable"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","result":"recovered"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Retries:   3,
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	var result map[string]any
	err := client.Call(context.Background(), "flickr.test.echo", nil, &result)
	if err != nil {
		t.Fatalf("expected call to succeed after retries, got error: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRESTSignsAuthenticatedCall(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:      "test-key",
		APISecret:   "test-secret",
		OAuthToken:  "test-token",
		OAuthSecret: "test-token-secret",
		HTTP:        server.Client(),
		Endpoints:   Endpoints{REST: server.URL + "/"},
	}

	err := client.Call(context.Background(), "flickr.test.login", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Authorization header starts with "OAuth "
	if !strings.HasPrefix(capturedAuth, "OAuth ") {
		t.Errorf("expected Authorization header to start with 'OAuth ', got %q", capturedAuth)
	}

	// Verify oauth_signature is present in the header
	if !strings.Contains(capturedAuth, "oauth_signature") {
		t.Errorf("expected Authorization header to contain 'oauth_signature', got %q", capturedAuth)
	}
}
