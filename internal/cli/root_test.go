package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestExecuteJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--json", "version"})

	err := Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected JSON output")
	}
}

func TestExecuteHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestFLICKREnvVars(t *testing.T) {
	// Create a temporary config file to point FLICKR_CONFIG at
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`schema_version: "2026-06-02"
default_profile: default
profiles:
  default:
    api_key: env-test-key
`), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	// Capture AppContext from PersistentPreRunE
	var capturedApp *AppContext
	testRoot := newRootCmd()
	testCmd := &cobra.Command{
		Use: "test-capture",
		RunE: func(cmd *cobra.Command, args []string) error {
			capturedApp = GetAppContext(cmd.Context())
			return nil
		},
	}
	testRoot.AddCommand(testCmd)

	// Test FLICKR_CONFIG
	t.Run("FLICKR_CONFIG", func(t *testing.T) {
		capturedApp = nil
		t.Setenv("FLICKR_CONFIG", cfgPath)
		t.Setenv("FLICKR_PROFILE", "") // clear any leftover

		buf := new(bytes.Buffer)
		testRoot.SetOut(buf)
		testRoot.SetArgs([]string{"test-capture"})
		if err := testRoot.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedApp == nil {
			t.Fatal("AppContext was not captured")
		}
		if capturedApp.ConfigFile != cfgPath {
			t.Errorf("expected ConfigFile=%s from FLICKR_CONFIG, got %s", cfgPath, capturedApp.ConfigFile)
		}
	})

	// Test FLICKR_PROFILE
	t.Run("FLICKR_PROFILE", func(t *testing.T) {
		capturedApp = nil
		t.Setenv("FLICKR_PROFILE", "staging")
		t.Setenv("FLICKR_CONFIG", "")

		buf := new(bytes.Buffer)
		testRoot.SetOut(buf)
		testRoot.SetArgs([]string{"test-capture"})
		if err := testRoot.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedApp == nil {
			t.Fatal("AppContext was not captured")
		}
		if capturedApp.Profile != "staging" {
			t.Errorf("expected Profile=staging from FLICKR_PROFILE, got %s", capturedApp.Profile)
		}
	})

	// Test that flags take precedence over env vars
	t.Run("flags_override_env", func(t *testing.T) {
		capturedApp = nil
		t.Setenv("FLICKR_CONFIG", "/should/not/be/used")
		t.Setenv("FLICKR_PROFILE", "should-not-be-used")

		buf := new(bytes.Buffer)
		testRoot.SetOut(buf)
		testRoot.SetArgs([]string{"--config", cfgPath, "--profile", "production", "test-capture"})
		if err := testRoot.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedApp == nil {
			t.Fatal("AppContext was not captured")
		}
		if capturedApp.ConfigFile != cfgPath {
			t.Errorf("expected flag to override env for ConfigFile, got %s", capturedApp.ConfigFile)
		}
		if capturedApp.Profile != "production" {
			t.Errorf("expected flag to override env for Profile, got %s", capturedApp.Profile)
		}
	})
}
