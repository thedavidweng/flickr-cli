package backup

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var unsafeChars = regexp.MustCompile(`[/\\` + "`" + `\x00-\x1f]`)

// SafeName sanitizes a string for use as a filename.
func SafeName(input string, fallback string) string {
	s := unsafeChars.ReplaceAllString(input, "_")
	s = strings.TrimRight(strings.TrimSpace(s), ".")

	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upper := strings.ToUpper(s)
	for _, r := range reserved {
		if upper == r {
			s = s + "_"
			break
		}
	}

	if s == "" {
		return fallback
	}

	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' || r == ' ' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	return result.String()
}

// UniqueName generates a unique filename to avoid conflicts.
func UniqueName(dir string, base string, photoID string) string {
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	fullPath := filepath.Join(dir, base)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return base
	}

	return fmt.Sprintf("%s-%s%s", name, photoID, ext)
}

// IDDirsPath generates the path for an id-dirs backup.
func IDDirsPath(dest string, photoID string, ext string) string {
	h := md5.New()
	io.WriteString(h, photoID)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	return filepath.Join(dest, hash[0:2], hash[2:4], photoID, photoID+"."+ext)
}

// AtomicWriteFile writes data to a temp file and then atomically renames it.
func AtomicWriteFile(path string, r io.Reader, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating dir: %w", err)
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("writing file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}

	return nil
}
