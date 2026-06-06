package flickr

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
)

// Size represents a photo size/variant.
type Size struct {
	Label  string `json:"label"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
	URL    string `json:"url"`
	Media  string `json:"media"`
}

// VideoStream represents a single video stream from flickr.video.getStreamInfo.
type VideoStream struct {
	Type   string `json:"type"` // e.g. "1080p", "720p", "orig"
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
}

// videoPriority maps stream type names to priority (lower = better).
var videoPriority = map[string]int{
	"orig":  0,
	"1080p": 1,
	"720p":  2,
	"360p":  3,
	"288p":  4,
	"700":   5,
	"300":   6,
	"100":   7,
}

// GetSizes returns available sizes for a photo.
func (c *Client) GetSizes(ctx context.Context, photoID string) ([]Size, error) {
	params := map[string]string{
		"photo_id": photoID,
	}

	var result struct {
		Sizes struct {
			Size []Size `json:"size"`
		} `json:"sizes"`
	}

	if err := c.Call(ctx, "flickr.photos.getSizes", params, &result); err != nil {
		return nil, err
	}

	return result.Sizes.Size, nil
}

// GetExif returns EXIF metadata for a photo.
func (c *Client) GetExif(ctx context.Context, photoID string) (*ExifData, error) {
	params := map[string]string{
		"photo_id": photoID,
	}

	var result ExifResponse
	if err := c.Call(ctx, "flickr.photos.getExif", params, &result); err != nil {
		return nil, err
	}

	return &result.Photo, nil
}

// GetVideoStreams returns available video streams for a photo.
// Requires authentication and a valid API key with video permissions.
func (c *Client) GetVideoStreams(ctx context.Context, photoID string) ([]VideoStream, error) {
	params := map[string]string{
		"photo_id": photoID,
	}

	var result struct {
		Streams struct {
			Stream []VideoStream `json:"stream"`
		} `json:"streams"`
	}

	if err := c.Call(ctx, "flickr.video.getStreamInfo", params, &result); err != nil {
		return nil, err
	}

	return result.Streams.Stream, nil
}

// SelectBestStream picks the highest quality video stream from available streams.
// Falls back to the best available if the preferred quality is not present.
func SelectBestStream(streams []VideoStream) (VideoStream, error) {
	if len(streams) == 0 {
		return VideoStream{}, fmt.Errorf("no video streams available")
	}

	sort.Slice(streams, func(i, j int) bool {
		pi, okI := videoPriority[streams[i].Type]
		pj, okJ := videoPriority[streams[j].Type]
		if !okI {
			pi = 99
		}
		if !okJ {
			pj = 99
		}
		return pi < pj
	})

	return streams[0], nil
}

// sizeCodeMap maps Flickr size codes to label patterns and max dimensions.
var sizeCodeMap = map[string]struct {
	labelContains string
	maxDim        int
}{
	"o":  {labelContains: "Original", maxDim: 0},
	"6k": {labelContains: "", maxDim: 6144},
	"5k": {labelContains: "", maxDim: 5120},
	"4k": {labelContains: "", maxDim: 4096},
	"3k": {labelContains: "", maxDim: 3072},
	"k":  {labelContains: "2048", maxDim: 2048},
	"h":  {labelContains: "1600", maxDim: 1600},
	"l":  {labelContains: "Large", maxDim: 1024},
	"c":  {labelContains: "800", maxDim: 800},
	"z":  {labelContains: "640", maxDim: 640},
	"m":  {labelContains: "Medium", maxDim: 500},
	"n":  {labelContains: "320", maxDim: 320},
	"s":  {labelContains: "Square", maxDim: 75},
	"q":  {labelContains: "Large Square", maxDim: 150},
	"t":  {labelContains: "Thumbnail", maxDim: 100},
}

// SelectSize picks the best matching size from available sizes.
// Supports legacy names (original, large, medium, small) and Flickr size codes (o, k, h, l, c, z, m, n, s, q, t).
func SelectSize(sizes []Size, wanted string) (Size, error) {
	if len(sizes) == 0 {
		return Size{}, fmt.Errorf("no sizes available")
	}

	// Check if it's a Flickr size code
	if info, ok := sizeCodeMap[wanted]; ok {
		return selectByCode(sizes, wanted, info)
	}

	// Legacy names
	switch wanted {
	case "original":
		for _, s := range sizes {
			if s.Label == "Original" {
				return s, nil
			}
		}
		return sizes[len(sizes)-1], nil

	case "large":
		for i := len(sizes) - 1; i >= 0; i-- {
			if sizes[i].Label != "Original" {
				return sizes[i], nil
			}
		}
		return sizes[len(sizes)-1], nil

	case "medium":
		for _, s := range sizes {
			if contains(s.Label, "Medium") {
				return s, nil
			}
		}
		for i := len(sizes) - 1; i >= 0; i-- {
			if sizes[i].Width <= 1024 {
				return sizes[i], nil
			}
		}
		return sizes[0], nil

	case "small":
		for _, s := range sizes {
			if contains(s.Label, "Small") {
				return s, nil
			}
		}
		for i := len(sizes) - 1; i >= 0; i-- {
			if sizes[i].Width <= 640 {
				return sizes[i], nil
			}
		}
		return sizes[0], nil

	default:
		return sizes[0], nil
	}
}

// selectByCode selects a size by Flickr size code.
func selectByCode(sizes []Size, code string, info struct {
	labelContains string
	maxDim        int
}) (Size, error) {
	// Try label match first
	if info.labelContains != "" {
		for _, s := range sizes {
			if contains(s.Label, info.labelContains) {
				return s, nil
			}
		}
	}

	// Try dimension match: find the largest size that fits within maxDim
	if info.maxDim > 0 {
		var best Size
		found := false
		for _, s := range sizes {
			if s.Width <= info.maxDim && s.Height <= info.maxDim {
				if !found || s.Width > best.Width {
					best = s
					found = true
				}
			}
		}
		if found {
			return best, nil
		}
	}

	// For "o" (original), fall back to largest
	if code == "o" {
		return sizes[len(sizes)-1], nil
	}

	// Fall back to closest match by width
	target := info.maxDim
	if target == 0 {
		target = 99999
	}
	best := sizes[0]
	bestDiff := abs(best.Width - target)
	for _, s := range sizes[1:] {
		diff := abs(s.Width - target)
		if diff < bestDiff {
			best = s
			bestDiff = diff
		}
	}
	return best, nil
}

// SelectSizeByMaxDimension selects the largest size whose width or height does not exceed maxPixels.
func SelectSizeByMaxDimension(sizes []Size, maxPixels int) (Size, error) {
	if len(sizes) == 0 {
		return Size{}, fmt.Errorf("no sizes available")
	}

	var best Size
	found := false
	for _, s := range sizes {
		if s.Width <= maxPixels && s.Height <= maxPixels {
			if !found || s.Width > best.Width {
				best = s
				found = true
			}
		}
	}
	if found {
		return best, nil
	}

	// All sizes exceed maxPixels; return the smallest
	return sizes[0], nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Flickr short URL base-58 alphabet (excludes 0, I, O, l to avoid confusion).
const base58Alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

// base58Decode decodes a base-58 encoded string to a uint64.
func base58Decode(s string) (uint64, error) {
	var result uint64
	for _, c := range s {
		idx := -1
		for i, a := range base58Alphabet {
			if a == c {
				idx = i
				break
			}
		}
		if idx < 0 {
			return 0, fmt.Errorf("invalid base58 character: %c", c)
		}
		result = result*58 + uint64(idx)
	}
	return result, nil
}

// DecodeShortURL resolves a flic.kr/p/XXXXX short URL to a numeric photo ID.
func DecodeShortURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Accept both flic.kr and www.flickr.com
	host := strings.ToLower(u.Host)
	if host != "flic.kr" && host != "flickr.com" && host != "www.flickr.com" {
		return "", fmt.Errorf("not a Flickr short URL: %s", rawURL)
	}

	// Extract the encoded ID from /p/XXXXX
	p := strings.TrimPrefix(u.Path, "/")
	parts := strings.SplitN(p, "/", 3)
	if len(parts) < 2 || parts[0] != "p" || parts[1] == "" {
		return "", fmt.Errorf("not a photo short URL: %s", rawURL)
	}

	id, err := base58Decode(parts[1])
	if err != nil {
		return "", fmt.Errorf("decoding short URL ID: %w", err)
	}

	return fmt.Sprintf("%d", id), nil
}

// ResolvePhotoID accepts a bare numeric ID, a full Flickr URL, or a short URL
// and returns the numeric photo ID.
func ResolvePhotoID(input string) (string, error) {
	s := strings.TrimSpace(input)

	// Bare numeric ID
	if _, err := fmt.Sscanf(s, "%d", new(uint64)); err == nil && !strings.Contains(s, "/") {
		return s, nil
	}

	// Short URL
	if strings.Contains(s, "flic.kr") {
		return DecodeShortURL(s)
	}

	// Full flickr.com URL: /photos/{user}/{photoID} or /photos/{user}/{photoID}/...
	if strings.Contains(s, "flickr.com") {
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		// Expected: ["photos", "{user}", "{photoID}", ...]
		if len(parts) >= 3 && parts[0] == "photos" {
			id := parts[2]
			if _, err := fmt.Sscanf(id, "%d", new(uint64)); err == nil {
				return id, nil
			}
		}
		return "", fmt.Errorf("cannot extract photo ID from URL: %s", s)
	}

	return "", fmt.Errorf("unrecognized photo identifier: %s", s)
}

// DeriveExtension determines the file extension from the download URL, media type,
// or original format field. Falls back to "jpg" for photos and "mp4" for videos.
func DeriveExtension(sourceURL string, media string, originalFormat string) string {
	// Try to extract from URL path (strip query parameters)
	if sourceURL != "" {
		if u, err := url.Parse(sourceURL); err == nil {
			ext := strings.TrimPrefix(strings.ToLower(path.Ext(u.Path)), ".")
			if ext != "" && len(ext) <= 5 {
				// Normalize common variations
				switch ext {
				case "jpeg":
					return "jpg"
				default:
					return ext
				}
			}
		}
	}

	// Try originalformat field from the API
	if originalFormat != "" {
		of := strings.ToLower(strings.TrimSpace(originalFormat))
		switch of {
		case "jpeg":
			return "jpg"
		default:
			return of
		}
	}

	// Default based on media type
	if media == "video" {
		return "mp4"
	}
	return "jpg"
}

// BestSizeURL returns the best available URL from a PhotoListItem's extras fields.
// Prefers original (url_o) > 2048 (url_k) > large (url_l) > medium (url_m) > small (url_s).
func BestSizeURL(p PhotoListItem) string {
	if p.URLO != "" {
		return p.URLO
	}
	if p.URLK != "" {
		return p.URLK
	}
	if p.URLL != "" {
		return p.URLL
	}
	if p.URLM != "" {
		return p.URLM
	}
	if p.URLS != "" {
		return p.URLS
	}
	return ""
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
