//go:build windows

package config

import "os"

// defaultStateDir returns the base directory for state files.
// On Windows, it uses %LOCALAPPDATA% (same as cache).
func defaultStateDir() string {
	return os.Getenv("LOCALAPPDATA")
}
