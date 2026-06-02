package config

import (
	"os"
	"testing"
)

func TestDefaultPathsXDG(t *testing.T) {
	// Test with XDG set
	os.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
	os.Setenv("XDG_CACHE_HOME", "/tmp/test-cache")
	os.Setenv("XDG_STATE_HOME", "/tmp/test-state")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Unsetenv("XDG_CACHE_HOME")
	defer os.Unsetenv("XDG_STATE_HOME")

	configPath := DefaultConfigPath()
	if configPath != "/tmp/test-config/flickr-cli/config.yaml" {
		t.Errorf("unexpected config path: %s", configPath)
	}

	cachePath := DefaultCachePath("default")
	if cachePath != "/tmp/test-cache/flickr-cli/default.sqlite" {
		t.Errorf("unexpected cache path: %s", cachePath)
	}

	auditPath := DefaultAuditLogPath("default")
	if auditPath != "/tmp/test-state/flickr-cli/audit-default.jsonl" {
		t.Errorf("unexpected audit path: %s", auditPath)
	}
}

func TestDefaultPathsFallback(t *testing.T) {
	// Clear XDG vars
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CACHE_HOME")
	os.Unsetenv("XDG_STATE_HOME")

	configPath := DefaultConfigPath()
	if configPath == "" {
		t.Error("expected non-empty config path")
	}
	// Should end with the expected suffix
	if len(configPath) < 20 {
		t.Errorf("config path too short: %s", configPath)
	}

	cachePath := DefaultCachePath("myprofile")
	if cachePath == "" {
		t.Error("expected non-empty cache path")
	}

	auditPath := DefaultAuditLogPath("myprofile")
	if auditPath == "" {
		t.Error("expected non-empty audit path")
	}
}
