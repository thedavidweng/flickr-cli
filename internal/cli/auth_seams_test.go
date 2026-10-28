package cli

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

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

	v, err := oobAuthorize("http://example.com/auth")
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

	_, err := oobAuthorize("http://example.com/auth")
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

// --- readFrom ---

func TestReadFromTrimmed(t *testing.T) {
	input := bytes.NewBufferString("  trimmed input  \n")
	got := readFrom(input)
	if got != "trimmed input" {
		t.Errorf("expected %q, got %q", "trimmed input", got)
	}
}
