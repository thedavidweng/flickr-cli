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
func Sync(ctx context.Context, db *DB, client *flickr.Client, opts SyncOptions) (*SyncResult, error) {
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

func syncAlbums(ctx context.Context, db *DB, client *flickr.Client) (int, error) {
	params := map[string]string{
		"user_id":  "me",
		"per_page": "500",
	}

	var result struct {
		Photosets struct {
			Photoset []struct {
				ID    string `json:"id"`
				Title struct {
					Content string `json:"_content"`
				} `json:"title"`
			} `json:"photoset"`
		} `json:"photosets"`
	}

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

func syncPhotos(ctx context.Context, db *DB, client *flickr.Client, limit int) (int, error) {
	perPage := 500
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	params := map[string]string{
		"user_id":  "me",
		"per_page": fmt.Sprintf("%d", perPage),
		"page":     "1",
	}

	var result struct {
		Photos struct {
			Photo []struct {
				ID string `json:"id"`
			} `json:"photo"`
			Pages int `json:"pages"`
		} `json:"photos"`
	}

	if err := client.Call(ctx, "flickr.people.getPhotos", params, &result); err != nil {
		return 0, err
	}

	count := 0
	for _, p := range result.Photos.Photo {
		payload, _ := json.Marshal(p)
		if err := db.UpsertPhoto(p.ID, string(payload)); err != nil {
			return count, err
		}
		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	return count, nil
}
