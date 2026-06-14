package upload

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

// AlbumResolver handles album lookup and creation during uploads.
type AlbumResolver struct {
	Client flickr.FlickrAPI
	cache  map[string]string // lowercase title -> id
	mu     sync.Mutex
}

// NewAlbumResolver creates a new AlbumResolver.
func NewAlbumResolver(client flickr.FlickrAPI) *AlbumResolver {
	return &AlbumResolver{
		Client: client,
		cache:  make(map[string]string),
	}
}

// Load fetches existing albums into the cache.
func (r *AlbumResolver) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	params := map[string]string{
		"user_id":  "me",
		"per_page": "500",
	}

	var result flickr.PhotosetListResponse

	if err := r.Client.Call(ctx, "flickr.photosets.getList", params, &result); err != nil {
		return fmt.Errorf("loading albums: %w", err)
	}

	for _, ps := range result.Photosets.Photoset {
		r.cache[strings.ToLower(ps.Title.Content)] = ps.ID
	}

	return nil
}

// ResolveOrCreate finds an album by title or creates it.
func (r *AlbumResolver) ResolveOrCreate(ctx context.Context, title, primaryPhotoID string) (id string, created bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(title))
	if key == "" {
		return "", false, fmt.Errorf("album title cannot be empty")
	}

	if existingID, ok := r.cache[key]; ok {
		return existingID, false, nil
	}

	// Create the album
	params := map[string]string{
		"title":            title,
		"primary_photo_id": primaryPhotoID,
	}

	var result struct {
		Photoset struct {
			ID string `json:"id"`
		} `json:"photoset"`
	}

	if err := r.Client.Call(ctx, "flickr.photosets.create", params, &result); err != nil {
		return "", false, fmt.Errorf("creating album: %w", err)
	}

	r.cache[key] = result.Photoset.ID
	return result.Photoset.ID, true, nil
}
