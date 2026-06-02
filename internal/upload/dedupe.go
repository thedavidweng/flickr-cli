package upload

import (
	"context"
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/checksum"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

// Deduplicator checks for existing photos by checksum machine tags.
type Deduplicator struct {
	Client    *flickr.Client
	Algorithm string
}

// CheckByChecksum searches Flickr for a photo with the given checksum machine tag.
func (d *Deduplicator) CheckByChecksum(ctx context.Context, hashValue string) (photoID string, found bool, err error) {
	mt := checksum.FormatMachineTag(d.Algorithm, hashValue)

	params := map[string]string{
		"user_id":      "me",
		"machine_tags": mt,
		"per_page":     "1",
	}

	var result struct {
		Photos struct {
			Photo []struct {
				ID string `json:"id"`
			} `json:"photo"`
			Total int `json:"total"`
		} `json:"photos"`
	}

	if err := d.Client.Call(ctx, "flickr.photos.search", params, &result); err != nil {
		return "", false, fmt.Errorf("searching for checksum: %w", err)
	}

	if result.Photos.Total > 0 && len(result.Photos.Photo) > 0 {
		return result.Photos.Photo[0].ID, true, nil
	}

	return "", false, nil
}
