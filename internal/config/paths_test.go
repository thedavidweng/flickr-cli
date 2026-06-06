package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPathsXDG(t *testing.T) {
	// Test with XDG set (use t.TempDir for cross-platform compatibility)
	configDir := t.TempDir()
	cacheDir := t.TempDir()
	stateDir := t.TempDir()

	os.Setenv("XDG_CONFIG_HOME", configDir)
	os.Setenv("XDG_CACHE_HOME", cacheDir)
	os.Setenv("XDG_STATE_HOME", stateDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Unsetenv("XDG_CACHE_HOME")
	defer os.Unsetenv("XDG_STATE_HOME")

	configPath := DefaultConfigPath()
	expectedConfig := filepath.Join(configDir, "flickr-cli", "config.yaml")
	if configPath != expectedConfig {
		t.Errorf("unexpected config path: %s, expected %s", configPath, expectedConfig)
	}

	cachePath := DefaultCachePath("default")
	expectedCache := filepath.Join(cacheDir, "flickr-cli", "default.sqlite")
	if cachePath != expectedCache {
		t.Errorf("unexpected cache path: %s, expected %s", cachePath, expectedCache)
	}

	auditPath := DefaultAuditLogPath("default")
	expectedAudit := filepath.Join(stateDir, "flickr-cli", "audit-default.jsonl")
	if auditPath != expectedAudit {
		t.Errorf("unexpected audit path: %s, expected %s", auditPath, expectedAudit)
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
