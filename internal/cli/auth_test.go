package cli

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/testutil"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func TestAuthHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"auth", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestAuthLoginHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"auth", "login", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestAuthStatusHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"auth", "status", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestAuthLogoutHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"auth", "logout", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestIsTerminal(t *testing.T) {
	result := isTerminal()
	_ = result
}

func TestReadLine(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		fmt.Fprintln(w, "test input")
		w.Close()
	}()

	result := readLine()
	if result != "test input" {
		t.Errorf("expected 'test input', got %q", result)
	}
}

func TestReadFrom(t *testing.T) {
	input := strings.NewReader("  hello world  \n")
	result := readFrom(input)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestLocalhostCallback(t *testing.T) {
	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	go func() {
		time.Sleep(100 * time.Millisecond)
		url := fmt.Sprintf("http://localhost:%d/?oauth_verifier=test-verifier", port)
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
		}
	}()

	verifier, err := localhostCallback(ctx, client, "test-token", "read", port)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verifier != "test-verifier" {
		t.Errorf("expected test-verifier, got %s", verifier)
	}
}

func TestLocalhostCallbackMissingVerifier(t *testing.T) {
	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	go func() {
		time.Sleep(100 * time.Millisecond)
		url := fmt.Sprintf("http://localhost:%d/", port)
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
		}
	}()

	_, err = localhostCallback(ctx, client, "test-token", "read", port)
	if err == nil {
		t.Error("expected error for missing verifier")
	}
}

func TestAuthLoginOOBNonInteractive(t *testing.T) {
	// Set up a fake Flickr server that handles request_token
	fake := testutil.NewFakeFlickr(t)
	defer fake.Server.Close()

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`schema_version: "2026-06-02"
default_profile: default
profiles:
  default:
    api_key: test-api-key
    api_secret: test-api-secret
    endpoints:
      request_token: %s/oauth/request_token
      authorize: %s/oauth/authorize
      access_token: %s/oauth/access_token
`, fake.Server.URL, fake.Server.URL, fake.Server.URL)
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Redirect os.Stdin to a pipe so isTerminal() returns false
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close() // close write end so readLine() gets EOF immediately

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Register the flags that authLoginCmd.RunE reads
	cmd.Flags().String("perms", "read", "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().String("callback", "oob", "")
	cmd.Flags().Int("callback-port", 0, "")
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	cmd.SetContext(WithAppContext(context.Background(), &AppContext{
		ConfigFile:  cfgPath,
		Profile:     "default",
		JSON:        true,
		Timeout:     30 * time.Second,
		Retries:     3,
		Concurrency: 4,
		RequestID:   uuid.New().String(),
		StartedAt:   time.Now(),
	}))

	err := authLoginCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for OOB in non-interactive mode")
	}

	// Parse the JSON output to verify the error
	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error body")
	}
	if !strings.Contains(env.Error.Message, "verifier required in non-interactive mode") {
		t.Errorf("expected 'verifier required in non-interactive mode', got: %s", env.Error.Message)
	}
}

func TestAuthLogoutDryRun(t *testing.T) {
	fake, cfgPath := setupFakeCLI(t)
	_ = fake

	// Capture original config content
	origContent, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading original config: %v", err)
	}

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	cmd.SetContext(WithAppContext(context.Background(), &AppContext{
		ConfigFile:  cfgPath,
		Profile:     "default",
		JSON:        true,
		DryRun:      true,
		Timeout:     30 * time.Second,
		Retries:     3,
		Concurrency: 4,
		RequestID:   uuid.New().String(),
		StartedAt:   time.Now(),
	}))

	err = authLogoutCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	if data["action"] != "clear_credentials" {
		t.Errorf("expected action=clear_credentials, got %v", data["action"])
	}

	// Verify config file was NOT modified
	afterContent, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading config after dry-run: %v", err)
	}
	if string(origContent) != string(afterContent) {
		t.Error("config file should not be modified during dry-run")
	}
}

func TestLocalhostCallbackContextCanceled(t *testing.T) {
	client := &flickr.Client{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	_, err = localhostCallback(ctx, client, "test-token", "read", port)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}
