//go:build !windows

package config

import (
	"os"
	"path/filepath"
)

// defaultStateDir returns the base directory for state files.
// On Unix, it checks XDG_STATE_HOME first, then falls back to ~/.local/state.
func defaultStateDir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "state")
}
