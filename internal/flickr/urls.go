package flickr

import (
	"context"
	"fmt"
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

// SelectSize picks the best matching size from available sizes.
func SelectSize(sizes []Size, wanted string) (Size, error) {
	if len(sizes) == 0 {
		return Size{}, fmt.Errorf("no sizes available")
	}

	switch wanted {
	case "original":
		for _, s := range sizes {
			if s.Label == "Original" {
				return s, nil
			}
		}
		// Fall back to largest
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
