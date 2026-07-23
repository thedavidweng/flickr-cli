package flickr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDetectOutboundIP(t *testing.T) {
	orig := oauthDial
	defer func() { oauthDial = orig }()

	oauthDial = func(network, addr string) (net.Conn, error) {
		pc, err := net.ListenPacket("udp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		udpConn, ok := pc.(*net.UDPConn)
		if !ok {
			return nil, fmt.Errorf("expected *net.UDPConn, got %T", pc)
		}
		return udpConn, nil
	}

	ip := detectOutboundIP()
	if ip == "" || ip == "localhost" {
		t.Errorf("expected real IP, got %q", ip)
	}
}

func TestDetectOutboundIPDialError(t *testing.T) {
	orig := oauthDial
	defer func() { oauthDial = orig }()

	oauthDial = func(network, addr string) (net.Conn, error) {
		return nil, fmt.Errorf("dial failed")
	}

	ip := detectOutboundIP()
	if ip != "localhost" {
		t.Errorf("expected localhost on dial error, got %q", ip)
	}
}

func TestLocalhostFlow(t *testing.T) {
	origListen := oauthListen
	defer func() { oauthListen = origListen }()
	oauthListen = net.Listen

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/request_token":
			_, _ = fmt.Fprint(w, "oauth_token=reqtok&oauth_token_secret=reqsecret&oauth_callback_confirmed=true")
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := &Client{
		APIKey: "test-key",
		HTTP:   server.Client(),
		Endpoints: Endpoints{
			REST:         server.URL + "/",
			RequestToken: server.URL + "/oauth/request_token",
			Authorize:    server.URL + "/oauth/authorize",
		},
	}

	flow := &OAuthFlow{Client: client}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// localhostFlow will bind to 127.0.0.1:0 and wait for a callback.
	// With a 2s timeout, it will expire — we're testing that it gets through
	// RequestToken and starts listening without error.
	_, _, err := flow.LocalhostFlow(ctx, "read", 0, func(authURL, callbackURL string) {})
	if err == nil {
		// If it somehow succeeded (callback arrived), that's fine too.
		return
	}
	// Expected: context.DeadlineExceeded since no callback arrives.
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestLocalhostFlowListenFallback(t *testing.T) {
	origListen := oauthListen
	defer func() { oauthListen = origListen }()

	callCount := 0
	oauthListen = func(network, addr string) (net.Listener, error) {
		callCount++
		if callCount == 1 {
			// First listen fails (simulates binding to outbound IP failing)
			return nil, fmt.Errorf("listen failed")
		}
		// Second listen (fallback to 127.0.0.1) succeeds
		return net.Listen(network, addr)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/request_token":
			_, _ = fmt.Fprint(w, "oauth_token=reqtok&oauth_token_secret=reqsecret&oauth_callback_confirmed=true")
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := &Client{
		APIKey: "test-key",
		HTTP:   server.Client(),
		Endpoints: Endpoints{
			REST:         server.URL + "/",
			RequestToken: server.URL + "/oauth/request_token",
			Authorize:    server.URL + "/oauth/authorize",
		},
	}

	flow := &OAuthFlow{Client: client}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, _, err := flow.LocalhostFlow(ctx, "read", 0, nil)
	if err == nil {
		return
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded after fallback, got %v", err)
	}
	if callCount < 2 {
		t.Errorf("expected fallback to second listen, got %d calls", callCount)
	}
}

func TestLocalhostFlowBothListenFail(t *testing.T) {
	origListen := oauthListen
	defer func() { oauthListen = origListen }()

	oauthListen = func(network, addr string) (net.Listener, error) {
		return nil, fmt.Errorf("listen failed")
	}

	client := &Client{APIKey: "test-key", HTTP: http.DefaultClient}

	flow := &OAuthFlow{Client: client}

	_, _, err := flow.LocalhostFlow(context.Background(), "read", 9999, nil)
	if err == nil {
		t.Fatal("expected error when both listens fail")
	}
}

func TestLocalhostFlowRequestTokenError(t *testing.T) {
	origListen := oauthListen
	defer func() { oauthListen = origListen }()
	oauthListen = net.Listen

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	client := &Client{
		APIKey: "test-key",
		HTTP:   server.Client(),
		Endpoints: Endpoints{
			REST:         server.URL + "/",
			RequestToken: server.URL + "/oauth/request_token",
		},
	}

	flow := &OAuthFlow{Client: client}

	_, _, err := flow.LocalhostFlow(context.Background(), "read", 0, nil)
	if err == nil {
		t.Fatal("expected error for RequestToken failure")
	}
}

func TestOOBRequestToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "oauth_token=reqtok&oauth_token_secret=reqsecret&oauth_callback_confirmed=true")
	}))
	defer server.Close()

	client := &Client{
		APIKey: "test-key",
		HTTP:   server.Client(),
		Endpoints: Endpoints{
			RequestToken: server.URL + "/oauth/request_token",
			Authorize:    server.URL + "/oauth/authorize",
		},
	}

	flow := &OAuthFlow{Client: client}

	authURL, reqToken, err := flow.OOBRequestToken(context.Background(), "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reqToken.Token != "reqtok" {
		t.Errorf("expected token reqtok, got %s", reqToken.Token)
	}
	if authURL == "" {
		t.Error("expected non-empty auth URL")
	}
}

// --- waitForCallback ---

func TestWaitForCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		addr, _ := ln.Addr().(*net.TCPAddr)
		url := fmt.Sprintf("http://localhost:%d/?oauth_verifier=test-verifier", addr.Port)
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	verifier, err := waitForCallback(ctx, ln)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verifier != "test-verifier" {
		t.Errorf("expected test-verifier, got %s", verifier)
	}
}

func TestWaitForCallbackMissingVerifier(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		addr, _ := ln.Addr().(*net.TCPAddr)
		url := fmt.Sprintf("http://localhost:%d/", addr.Port)
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
		// Cancel context after the no-verifier request returns 200
		cancel()
	}()

	_, err = waitForCallback(ctx, ln)
	if err == nil {
		t.Error("expected error for missing verifier")
	}
}

func TestWaitForCallbackContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}

	_, err = waitForCallback(ctx, ln)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}
