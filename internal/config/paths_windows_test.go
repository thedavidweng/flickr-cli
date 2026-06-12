//go:build windows

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPathsWindows(t *testing.T) {
	// Ensure XDG vars are not set
	_ = os.Unsetenv("XDG_CONFIG_HOME")
	_ = os.Unsetenv("XDG_CACHE_HOME")
	_ = os.Unsetenv("XDG_STATE_HOME")

	configPath := DefaultConfigPath()
	cachePath := DefaultCachePath("default")
	auditPath := DefaultAuditLogPath("default")

	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	if appData != "" {
		expectedConfig := filepath.Join(appData, "flickr-cli", "config.yaml")
		if configPath != expectedConfig {
			t.Errorf("config path: got %s, want %s", configPath, expectedConfig)
		}
	}

	if localAppData != "" {
		expectedCache := filepath.Join(localAppData, "flickr-cli", "default.sqlite")
		if cachePath != expectedCache {
			t.Errorf("cache path: got %s, want %s", cachePath, expectedCache)
		}

		expectedAudit := filepath.Join(localAppData, "flickr-cli", "audit-default.jsonl")
		if auditPath != expectedAudit {
			t.Errorf("audit path: got %s, want %s", auditPath, expectedAudit)
		}
	}
}

func TestDefaultPathsWindowsXDGOverride(t *testing.T) {
	configDir := t.TempDir()
	cacheDir := t.TempDir()
	stateDir := t.TempDir()

	_ = os.Setenv("XDG_CONFIG_HOME", configDir)
	_ = os.Setenv("XDG_CACHE_HOME", cacheDir)
	_ = os.Setenv("XDG_STATE_HOME", stateDir)
	defer func() { _ = os.Unsetenv("XDG_CONFIG_HOME") }()
	defer func() { _ = os.Unsetenv("XDG_CACHE_HOME") }()
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	configPath := DefaultConfigPath()
	if !strings.HasPrefix(configPath, configDir) {
		t.Errorf("expected XDG override for config, got %s", configPath)
	}

	cachePath := DefaultCachePath("default")
	if !strings.HasPrefix(cachePath, cacheDir) {
		t.Errorf("expected XDG override for cache, got %s", cachePath)
	}

	auditPath := DefaultAuditLogPath("default")
	if !strings.HasPrefix(auditPath, stateDir) {
		t.Errorf("expected XDG override for audit, got %s", auditPath)
	}
}
