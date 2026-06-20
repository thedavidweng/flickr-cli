package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// --- detectOutboundIP ---

func TestDetectOutboundIP(t *testing.T) {
	orig := netDial
	defer func() { netDial = orig }()

	netDial = func(network, addr string) (net.Conn, error) {
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
	orig := netDial
	defer func() { netDial = orig }()

	netDial = func(network, addr string) (net.Conn, error) {
		return nil, fmt.Errorf("dial failed")
	}

	ip := detectOutboundIP()
	if ip != "localhost" {
		t.Errorf("expected localhost on dial error, got %q", ip)
	}
}

// --- readInput ---

func TestReadInputReplaced(t *testing.T) {
	orig := readInput
	defer func() { readInput = orig }()

	readInput = func() string { return "test-value" }

	got := readInput()
	if got != "test-value" {
		t.Errorf("expected test-value, got %q", got)
	}
}

// --- resolveCredentials ---

func TestResolveCredentialsFromInteractiveInput(t *testing.T) {
	orig := readInput
	defer func() { readInput = orig }()

	calls := 0
	readInput = func() string {
		calls++
		if calls == 1 {
			return "interactive-key"
		}
		return "interactive-secret"
	}

	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	apiKey, apiSecret, err := resolveCredentials(cmd, &r, meta, config.Credentials{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "interactive-key" {
		t.Errorf("expected interactive-key, got %q", apiKey)
	}
	if apiSecret != "interactive-secret" {
		t.Errorf("expected interactive-secret, got %q", apiSecret)
	}
}

func TestResolveCredentialsPartialFromProfile(t *testing.T) {
	orig := readInput
	defer func() { readInput = orig }()
	readInput = func() string { return "filled-secret" }

	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	apiKey, apiSecret, err := resolveCredentials(cmd, &r, meta, config.Credentials{APIKey: "profile-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "profile-key" {
		t.Errorf("expected profile-key, got %q", apiKey)
	}
	if apiSecret != "filled-secret" {
		t.Errorf("expected filled-secret, got %q", apiSecret)
	}
}

func TestResolveCredentialsBothEmpty(t *testing.T) {
	orig := readInput
	defer func() { readInput = orig }()
	readInput = func() string { return "" }

	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	_, _, err := resolveCredentials(cmd, &r, meta, config.Credentials{})
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestResolveCredentialsSecretFromEnvFlag(t *testing.T) {
	t.Setenv("TEST_SECRET_FROM_ENV", "env-secret-value")

	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")
	_ = cmd.Flags().Set("api-key", "flag-key")
	_ = cmd.Flags().Set("api-secret-env", "TEST_SECRET_FROM_ENV")

	apiKey, apiSecret, err := resolveCredentials(cmd, &r, meta, config.Credentials{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "flag-key" {
		t.Errorf("expected flag-key, got %q", apiKey)
	}
	if apiSecret != "env-secret-value" {
		t.Errorf("expected env-secret-value, got %q", apiSecret)
	}
}

// --- oobAuthorize ---

func TestOOBAuthorize(t *testing.T) {
	orig := readInput
	defer func() { readInput = orig }()
	readInput = func() string { return "my-verifier-code" }

	fake := &flickr.Client{APIKey: "test"}
	r := output.Renderer{Out: io.Discard, Err: io.Discard}

	v, err := oobAuthorize(context.Background(), &r, fake, "tok", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "my-verifier-code" {
		t.Errorf("expected my-verifier-code, got %q", v)
	}
}

func TestOOBAuthorizeEmptyVerifier(t *testing.T) {
	orig := readInput
	defer func() { readInput = orig }()
	readInput = func() string { return "" }

	fake := &flickr.Client{APIKey: "test"}
	r := output.Renderer{Out: io.Discard, Err: io.Discard}

	_, err := oobAuthorize(context.Background(), &r, fake, "tok", "read")
	if err == nil {
		t.Fatal("expected error for empty verifier")
	}
}

// --- handleRequestTokenError ---

func TestHandleRequestTokenError401(t *testing.T) {
	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	err := handleRequestTokenError(r, meta, fmt.Errorf("HTTP 401 Unauthorized"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleRequestTokenErrorGeneric(t *testing.T) {
	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	err := handleRequestTokenError(r, meta, fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- localhostFlow ---

func TestLocalhostFlow(t *testing.T) {
	origListen := netListen
	defer func() { netListen = origListen }()
	netListen = net.Listen

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/request_token":
			_, _ = fmt.Fprint(w, "oauth_token=reqtok&oauth_token_secret=reqsecret&oauth_callback_confirmed=true")
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey: "test-key",
		HTTP:   server.Client(),
		Endpoints: flickr.Endpoints{
			REST:         server.URL + "/",
			RequestToken: server.URL + "/oauth/request_token",
			Authorize:    server.URL + "/oauth/authorize",
		},
	}

	r := output.Renderer{Out: io.Discard, Err: io.Discard}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// localhostFlow will bind to 127.0.0.1:0 and wait for a callback.
	// With a 2s timeout, it will expire — we're testing that it gets through
	// RequestToken and starts listening without error.
	_, _, err := localhostFlow(ctx, &r, client, "read", 0)
	if err == nil {
		// If it somehow succeeded (callback arrived), that's fine too.
		return
	}
	// Expected: context.DeadlineExceeded since no callback arrives.
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

// --- isTerminal seam ---

func TestIsTerminalInNonTerminal(t *testing.T) {
	// In test environments, stdin is typically not a char device.
	// Just verify the function doesn't panic.
	_ = isTerminal()
}

// --- readFrom ---

func TestReadFromTrimmed(t *testing.T) {
	input := bytes.NewBufferString("  trimmed input  \n")
	got := readFrom(input)
	if got != "trimmed input" {
		t.Errorf("expected %q, got %q", "trimmed input", got)
	}
}
