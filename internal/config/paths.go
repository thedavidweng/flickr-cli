package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default config file path.
// On Unix, XDG_CONFIG_HOME takes precedence; otherwise os.UserConfigDir() is used.
// On Windows, this resolves to %APPDATA%\flickr-cli\config.yaml.
func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "flickr-cli", "config.yaml")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".config", "flickr-cli", "config.yaml")
	}
	return filepath.Join(dir, "flickr-cli", "config.yaml")
}

// DefaultCachePath returns the default cache database path for a profile.
// On Unix, XDG_CACHE_HOME takes precedence; otherwise os.UserCacheDir() is used.
// On Windows, this resolves to %LOCALAPPDATA%\flickr-cli\{profile}.sqlite.
func DefaultCachePath(profile string) string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "flickr-cli", profile+".sqlite")
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(".cache", "flickr-cli", profile+".sqlite")
	}
	return filepath.Join(dir, "flickr-cli", profile+".sqlite")
}

// DefaultAuditLogPath returns the default audit log path for a profile.
// On Unix, XDG_STATE_HOME takes precedence; otherwise ~/.local/state is used.
// On Windows, this resolves to %LOCALAPPDATA%\flickr-cli\audit-{profile}.jsonl.
func DefaultAuditLogPath(profile string) string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "flickr-cli", "audit-"+profile+".jsonl")
	}
	dir := defaultStateDir()
	if dir == "" {
		return filepath.Join(".local", "state", "flickr-cli", "audit-"+profile+".jsonl")
	}
	return filepath.Join(dir, "flickr-cli", "audit-"+profile+".jsonl")
}
