package backup

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var unsafeChars = regexp.MustCompile(`[/\\` + "`" + `\x00-\x1f]`)

// SafeName sanitizes a string for use as a filename.
func SafeName(input, fallback string) string {
	s := unsafeChars.ReplaceAllString(input, "_")
	s = strings.TrimRight(strings.TrimSpace(s), ".")

	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upper := strings.ToUpper(s)
	for _, r := range reserved {
		if upper == r {
			s += "_"
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

// IDDirsPath generates the path for an id-dirs backup.
func IDDirsPath(dest, photoID, ext string) string {
	h := md5.New()
	_, _ = io.WriteString(h, photoID)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	return filepath.Join(dest, hash[0:2], hash[2:4], photoID, photoID+"."+ext)
}
