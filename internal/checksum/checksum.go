package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// FileHash computes the hash of a file using the specified algorithm.
func FileHash(path string, algorithm string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	if err := ValidateAlgorithm(algorithm); err != nil {
		return "", err
	}

	var h hash.Hash
	switch algorithm {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	}

	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ValidateAlgorithm checks if the hash algorithm is supported.
func ValidateAlgorithm(algorithm string) error {
	switch algorithm {
	case "md5", "sha1":
		return nil
	default:
		return fmt.Errorf("unsupported hash algorithm: %s (supported: md5, sha1)", algorithm)
	}
}
