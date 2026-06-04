package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

// SyncOptions configures cache sync.
type SyncOptions struct {
	Albums bool
	Photos bool
	Limit  int
}

// SyncResult is the result of a cache sync.
type SyncResult struct {
	AlbumsSynced int `json:"albums_synced"`
	PhotosSynced int `json:"photos_synced"`
}

// Sync synchronizes the cache with Flickr.
func Sync(ctx context.Context, db *DB, client flickr.FlickrAPI, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{}

	if opts.Albums {
		count, err := syncAlbums(ctx, db, client)
		if err != nil {
			return nil, fmt.Errorf("syncing albums: %w", err)
		}
		result.AlbumsSynced = count
	}

	if opts.Photos {
		count, err := syncPhotos(ctx, db, client, opts.Limit)
		if err != nil {
			return nil, fmt.Errorf("syncing photos: %w", err)
		}
		result.PhotosSynced = count
	}

	return result, nil
}

func syncAlbums(ctx context.Context, db *DB, client flickr.FlickrAPI) (int, error) {
	params := map[string]string{
		"user_id":  "me",
		"per_page": "500",
	}

	var result flickr.PhotosetListResponse

	if err := client.Call(ctx, "flickr.photosets.getList", params, &result); err != nil {
		return 0, err
	}

	count := 0
	for _, ps := range result.Photosets.Photoset {
		payload, _ := json.Marshal(ps)
		if err := db.UpsertAlbum(ps.ID, ps.Title.Content, string(payload)); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

func syncPhotos(ctx context.Context, db *DB, client flickr.FlickrAPI, limit int) (int, error) {
	perPage := 500
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	count := 0
	page := 1

	for {
		params := map[string]string{
			"user_id":  "me",
			"per_page": fmt.Sprintf("%d", perPage),
			"page":     fmt.Sprintf("%d", page),
			"extras":   flickr.DefaultExtras,
		}

		var result flickr.PhotoListResponse

		if err := client.Call(ctx, "flickr.people.getPhotos", params, &result); err != nil {
			return count, err
		}

		for _, p := range result.Photos.Photo {
			payload, _ := json.Marshal(p)
			if err := db.UpsertPhoto(p.ID, string(payload)); err != nil {
				return count, err
			}
			count++
			if limit > 0 && count >= limit {
				return count, nil
			}
		}

		if result.Photos.Pages == 0 || page >= result.Photos.Pages {
			break
		}
		page++
	}

	return count, nil
}
