package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default config file path following XDG conventions.
func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "flickr-cli", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "flickr-cli", "config.yaml")
	}
	return filepath.Join(home, ".config", "flickr-cli", "config.yaml")
}

// DefaultCachePath returns the default cache database path for a profile.
func DefaultCachePath(profile string) string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "flickr-cli", profile+".sqlite")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", "flickr-cli", profile+".sqlite")
	}
	return filepath.Join(home, ".cache", "flickr-cli", profile+".sqlite")
}

// DefaultAuditLogPath returns the default audit log path for a profile.
func DefaultAuditLogPath(profile string) string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "flickr-cli", "audit-"+profile+".jsonl")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".local", "state", "flickr-cli", "audit-"+profile+".jsonl")
	}
	return filepath.Join(home, ".local", "state", "flickr-cli", "audit-"+profile+".jsonl")
}
