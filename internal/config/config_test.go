package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CurrentProfile != "default" {
		t.Errorf("expected current_profile=default, got %s", cfg.CurrentProfile)
	}

	// Verify file was created with secure permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}

	// Load again - should work
	cfg2, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if cfg2.CurrentProfile != cfg.CurrentProfile {
		t.Errorf("current_profile mismatch")
	}
}

func TestSavePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentProfile: "default",
		Profiles: map[string]*Profile{
			"default": {
				APIKey:    "test-key",
				APISecret: "test-secret",
			},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestGetProfile(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			"default": {APIKey: "key1"},
			"work":    {APIKey: "key2"},
		},
	}

	p, err := cfg.GetProfile("work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.APIKey != "key2" {
		t.Errorf("expected key2, got %s", p.APIKey)
	}

	_, err = cfg.GetProfile("nonexistent")
	if err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestSetProfile(t *testing.T) {
	cfg := &Config{}
	cfg.SetProfile("test", &Profile{APIKey: "key"})
	if cfg.Profiles["test"].APIKey != "key" {
		t.Error("expected profile to be set")
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte("invalid: yaml: content: {"), 0o644)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := &Config{CurrentProfile: "default"}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Error("config file should exist")
	}
}
