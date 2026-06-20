package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

func TestOOBAuthorizeSuccess(t *testing.T) {
	oldTerminal := isTerminal
	oldRead := readLine
	defer func() {
		isTerminal = oldTerminal
		readLine = oldRead
	}()

	isTerminal = func() bool { return true }
	readLine = func() string { return "my-verifier" }

	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}

	fake, _ := setupFakeCLI(t)

	verifier, err := oobAuthorize(t.Context(), r, fake.Client(), "req-token", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verifier != "my-verifier" {
		t.Errorf("expected 'my-verifier', got %q", verifier)
	}
}

func TestOOBAuthorizeEmptyVerifier(t *testing.T) {
	oldTerminal := isTerminal
	oldRead := readLine
	defer func() {
		isTerminal = oldTerminal
		readLine = oldRead
	}()

	isTerminal = func() bool { return true }
	readLine = func() string { return "" }

	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}

	fake, _ := setupFakeCLI(t)

	_, err := oobAuthorize(t.Context(), r, fake.Client(), "req-token", "read")
	if err == nil {
		t.Fatal("expected error for empty verifier")
	}
}

func TestResolveCredentialsFromInteractiveInput(t *testing.T) {
	oldTerminal := isTerminal
	oldRead := readLine
	defer func() {
		isTerminal = oldTerminal
		readLine = oldRead
	}()

	isTerminal = func() bool { return true }
	calls := 0
	readLine = func() string {
		calls++
		if calls == 1 {
			return "interactive-key"
		}
		return "interactive-secret"
	}

	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "auth.login"}
	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	apiKey, apiSecret, err := resolveCredentials(cmd, r, meta, config.Credentials{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "interactive-key" {
		t.Errorf("expected 'interactive-key', got %q", apiKey)
	}
	if apiSecret != "interactive-secret" {
		t.Errorf("expected 'interactive-secret', got %q", apiSecret)
	}
}

func TestResolveCredentialsPartialFromProfile(t *testing.T) {
	oldTerminal := isTerminal
	oldRead := readLine
	defer func() {
		isTerminal = oldTerminal
		readLine = oldRead
	}()

	isTerminal = func() bool { return true }
	readLine = func() string { return "prompted-secret" }

	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "auth.login"}
	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	apiKey, apiSecret, err := resolveCredentials(cmd, r, meta, config.Credentials{
		APIKey: "profile-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "profile-key" {
		t.Errorf("expected 'profile-key', got %q", apiKey)
	}
	if apiSecret != "prompted-secret" {
		t.Errorf("expected 'prompted-secret', got %q", apiSecret)
	}
}

func TestResolveCredentialsEmptyFlags(t *testing.T) {
	oldTerminal := isTerminal
	oldRead := readLine
	defer func() {
		isTerminal = oldTerminal
		readLine = oldRead
	}()

	isTerminal = func() bool { return true }
	readLine = func() string { return "" }

	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "auth.login"}
	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")

	_, _, err := resolveCredentials(cmd, r, meta, config.Credentials{})
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestHandleRequestTokenError(t *testing.T) {
	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "auth.login"}

	tests := []struct {
		name   string
		errMsg string
	}{
		{"401 error", "status 401"},
		{"400 error", "status 400"},
		{"generic error", "connection timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleRequestTokenError(r, meta, fmt.Errorf("%s", tt.errMsg))
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestResolveCredentialsFromFlags(t *testing.T) {
	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "auth.login"}
	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")
	_ = cmd.Flags().Set("api-key", "flag-key")
	_ = cmd.Flags().Set("api-secret", "flag-secret")

	apiKey, apiSecret, err := resolveCredentials(cmd, r, meta, config.Credentials{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "flag-key" {
		t.Errorf("expected 'flag-key', got %q", apiKey)
	}
	if apiSecret != "flag-secret" {
		t.Errorf("expected 'flag-secret', got %q", apiSecret)
	}
}

func TestOOBAuthorizeNonInteractive(t *testing.T) {
	oldTerminal := isTerminal
	defer func() { isTerminal = oldTerminal }()
	isTerminal = func() bool { return false }

	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}

	fake, _ := setupFakeCLI(t)

	_, err := oobAuthorize(t.Context(), r, fake.Client(), "req-token", "read")
	if err == nil {
		t.Fatal("expected error for non-interactive OOB")
	}
}

func TestLocalhostFlowTimeout(t *testing.T) {
	// Verify localhostFlow respects context cancellation.
	// We can't fully test it (requires browser callback), but we can
	// verify it returns an error when context is canceled.
	buf := new(bytes.Buffer)
	r := &output.Renderer{Out: buf, Err: buf, JSON: true}

	fake, _ := setupFakeCLI(t)
	client := fake.Client()

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, _, err := localhostFlow(ctx, r, client, "read", 0)
	if err == nil {
		t.Fatal("expected error for timed-out localhost flow")
	}
}
