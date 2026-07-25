package flickr

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
)

type Size struct {
	Label  string `json:"label"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
	URL    string `json:"url"`
	Media  string `json:"media"`
}

type VideoStream struct {
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
}

// videoPriority orders stream types; lower is higher quality.
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

// GetVideoStreams requires an API key with video permissions.
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
	"l":  {labelContains: "1024", maxDim: 1024},
	"c":  {labelContains: "800", maxDim: 800},
	"z":  {labelContains: "640", maxDim: 640},
	"m":  {labelContains: "Medium", maxDim: 500},
	"n":  {labelContains: "320", maxDim: 320},
	"s":  {labelContains: "Square", maxDim: 75},
	"q":  {labelContains: "Large Square", maxDim: 150},
	"t":  {labelContains: "Thumbnail", maxDim: 100},
}

// SelectSize accepts legacy names (original, large, medium, small) or Flickr
// size codes (o, k, h, l, c, z, m, n, s, q, t).
func SelectSize(sizes []Size, wanted string) (Size, error) {
	if len(sizes) == 0 {
		return Size{}, fmt.Errorf("no sizes available")
	}

	if info, ok := sizeCodeMap[wanted]; ok {
		return selectByCode(sizes, wanted, info)
	}

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
			if strings.Contains(s.Label, "Medium") {
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
			if strings.Contains(s.Label, "Small") {
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

func selectByCode(sizes []Size, code string, info struct {
	labelContains string
	maxDim        int
}) (Size, error) {
	if info.labelContains != "" {
		for _, s := range sizes {
			if strings.Contains(s.Label, info.labelContains) {
				return s, nil
			}
		}
	}

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

	if code == "o" {
		return sizes[len(sizes)-1], nil
	}

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

	return sizes[0], nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// base58Alphabet is Flickr's short-URL alphabet; it excludes 0, I, O, and l.
const base58Alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var base58Index = func() map[rune]uint64 {
	m := make(map[rune]uint64, len(base58Alphabet))
	for i, c := range base58Alphabet {
		m[c] = uint64(i)
	}
	return m
}()

func base58Decode(s string) (uint64, error) {
	var result uint64
	for _, c := range s {
		idx, ok := base58Index[c]
		if !ok {
			return 0, fmt.Errorf("invalid base58 character: %c", c)
		}
		result = result*58 + idx
	}
	return result, nil
}

func DecodeShortURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	host := strings.ToLower(u.Host)
	if host != "flic.kr" && host != "flickr.com" && host != "www.flickr.com" {
		return "", fmt.Errorf("not a Flickr short URL: %s", rawURL)
	}

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

// ResolvePhotoID accepts a bare numeric ID, a full Flickr URL, or a short URL.
func ResolvePhotoID(input string) (string, error) {
	s := strings.TrimSpace(input)

	if _, err := fmt.Sscanf(s, "%d", new(uint64)); err == nil && !strings.Contains(s, "/") {
		return s, nil
	}

	if strings.Contains(s, "flic.kr") {
		return DecodeShortURL(s)
	}

	if strings.Contains(s, "flickr.com") {
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
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

// DeriveExtension resolves the file extension from the URL, media type, or
// original format, defaulting to jpg for photos and mp4 for videos.
func DeriveExtension(sourceURL, media, originalFormat string) string {
	if sourceURL != "" {
		if u, err := url.Parse(sourceURL); err == nil {
			ext := strings.TrimPrefix(strings.ToLower(path.Ext(u.Path)), ".")
			if ext != "" && len(ext) <= 5 {
				switch ext {
				case "jpeg":
					return "jpg"
				default:
					return ext
				}
			}
		}
	}

	if originalFormat != "" {
		of := strings.ToLower(strings.TrimSpace(originalFormat))
		switch of {
		case "jpeg":
			return "jpg"
		default:
			return of
		}
	}

	if media == "video" {
		return "mp4"
	}
	return "jpg"
}

// BestSizeURL prefers url_o > url_k > url_l > url_m > url_s.
func BestSizeURL(p *PhotoListItem) string {
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
