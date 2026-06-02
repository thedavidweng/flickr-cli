package piwigo

import (
	"path/filepath"
	"strings"
)

// PrivacyFromLevel maps Piwigo privacy levels to Flickr privacy strings.
func PrivacyFromLevel(level int) string {
	switch {
	case level == 0:
		return "public"
	case level >= 1 && level <= 4:
		return "friends-family"
	default:
		return "private"
	}
}

// LocalPath resolves a Piwigo image path to a local file path.
func LocalPath(uploadsRoot string, piwigoPath string) string {
	p := piwigoPath
	for _, prefix := range []string{"./upload/", "upload/", "/upload/"} {
		if strings.HasPrefix(p, prefix) {
			p = p[len(prefix):]
			break
		}
	}

	fullPath := filepath.Join(uploadsRoot, p)
	return filepath.Clean(fullPath)
}

// Tags builds the tag list for an image record.
func Tags(record ImageRecord, hashAlg string, hashValue string) []string {
	tags := make([]string, 0, len(record.Tags)+1)
	tags = append(tags, record.Tags...)

	if hashAlg != "" && hashValue != "" {
		tags = append(tags, "checksum:"+hashAlg+"="+hashValue)
	}

	return tags
}

// Albums builds the album list for an image record.
func Albums(record ImageRecord, prefix string, importAlbum string) []string {
	var albums []string

	if importAlbum != "" {
		albums = append(albums, importAlbum)
	}

	for _, cat := range record.Categories {
		if cat == "" {
			continue
		}
		if prefix != "" {
			albums = append(albums, prefix+cat)
		} else {
			albums = append(albums, cat)
		}
	}

	return albums
}
