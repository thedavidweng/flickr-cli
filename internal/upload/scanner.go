package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultAcceptedExts lists the default accepted file extensions.
var DefaultAcceptedExts = []string{
	"jpg", "jpeg", "png", "gif", "tif", "tiff", "bmp", "webp",
	"heic", "heif", "mp4", "mov", "avi", "m4v",
}

// ScanOptions configures the file scanner.
type ScanOptions struct {
	Recursive   bool
	AcceptedExt []string
}

// LocalFile represents a file found by the scanner.
type LocalFile struct {
	Path    string
	RelPath string
	Name    string
	Ext     string
	Size    int64
}

// InvalidFile represents a file with an unaccepted extension.
type InvalidFile struct {
	Path string
	Ext  string
}

// Scan finds valid files in the given paths.
func Scan(paths []string, opts ScanOptions) (valid []LocalFile, invalid []InvalidFile, err error) {
	accepted := make(map[string]bool)
	for _, ext := range opts.AcceptedExt {
		accepted[strings.ToLower(ext)] = true
	}
	if len(accepted) == 0 {
		for _, ext := range DefaultAcceptedExts {
			accepted[ext] = true
		}
	}

	for _, p := range paths {
		info, err := os.Lstat(p)
		if err != nil {
			return nil, nil, fmt.Errorf("accessing %s: %w", p, err)
		}

		if info.Mode().IsRegular() {
			lf, inv := classifyFile(p, p, info, accepted)
			if lf != nil {
				valid = append(valid, *lf)
			} else if inv != nil {
				invalid = append(invalid, *inv)
			}
			continue
		}

		if info.IsDir() {
			if opts.Recursive {
				err := filepath.Walk(p, func(path string, fi os.FileInfo, walkErr error) error {
					if walkErr != nil {
						return walkErr
					}
					if fi.Mode().IsRegular() {
						rel, _ := filepath.Rel(p, path)
						lf, inv := classifyFile(path, rel, fi, accepted)
						if lf != nil {
							valid = append(valid, *lf)
						} else if inv != nil {
							invalid = append(invalid, *inv)
						}
					}
					return nil
				})
				if err != nil {
					return nil, nil, fmt.Errorf("walking %s: %w", p, err)
				}
			} else {
				entries, err := os.ReadDir(p)
				if err != nil {
					return nil, nil, fmt.Errorf("reading dir %s: %w", p, err)
				}
				for _, entry := range entries {
					if !entry.IsDir() {
						fullPath := filepath.Join(p, entry.Name())
						fi, err := entry.Info()
						if err != nil {
							continue
						}
						lf, inv := classifyFile(fullPath, entry.Name(), fi, accepted)
						if lf != nil {
							valid = append(valid, *lf)
						} else if inv != nil {
							invalid = append(invalid, *inv)
						}
					}
				}
			}
		}
	}

	// Sort by path for stable output
	sort.Slice(valid, func(i, j int) bool {
		return valid[i].Path < valid[j].Path
	})
	sort.Slice(invalid, func(i, j int) bool {
		return invalid[i].Path < invalid[j].Path
	})

	return valid, invalid, nil
}

func classifyFile(path, relPath string, info os.FileInfo, accepted map[string]bool) (*LocalFile, *InvalidFile) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if accepted[ext] {
		return &LocalFile{
			Path:    path,
			RelPath: relPath,
			Name:    info.Name(),
			Ext:     ext,
			Size:    info.Size(),
		}, nil
	}
	return nil, &InvalidFile{Path: path, Ext: ext}
}
