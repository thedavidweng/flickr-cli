package backup

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

// PlanMode defines the backup mode.
type PlanMode string

const (
	BackupAlbums PlanMode = "albums"
	BackupUser   PlanMode = "user"
	BackupIDDirs PlanMode = "id_dirs"
)

// BackupPlanOptions configures the backup planner.
type BackupPlanOptions struct {
	Mode        PlanMode
	Dest        string
	AlbumTitles []string
	AlbumIDs    []string
	All         bool
	UserID      string
	Size        string
	SizeMax     int
	Metadata    string
	Force       bool
	Exif        bool
}

// BackupPlan contains the items to back up.
type BackupPlan struct {
	Items    []BackupItem `json:"items"`
	Warnings []string     `json:"warnings,omitempty"`
}

// BackupItem is a single photo to back up.
type BackupItem struct {
	PhotoID        string `json:"photo_id"`
	Title          string `json:"title,omitempty"`
	AlbumID        string `json:"album_id,omitempty"`
	AlbumName      string `json:"album_name,omitempty"`
	Media          string `json:"media,omitempty"`
	OriginalFormat string `json:"original_format,omitempty"`
	URLO           string `json:"url_o,omitempty"`
	URLK           string `json:"url_k,omitempty"`
	Secret         string `json:"secret,omitempty"`
}

// BuildPlan creates a backup plan from the given options.
func BuildPlan(ctx context.Context, client flickr.FlickrAPI, opts *BackupPlanOptions) (*BackupPlan, error) {
	plan := &BackupPlan{}

	switch opts.Mode {
	case BackupAlbums:
		return buildAlbumPlan(ctx, client, opts, plan)
	case BackupUser:
		return buildUserPlan(ctx, client, opts, plan)
	case BackupIDDirs:
		return buildIDDirsPlan(ctx, client, opts, plan)
	default:
		return nil, fmt.Errorf("unsupported backup mode: %s", opts.Mode)
	}
}

func buildAlbumPlan(ctx context.Context, client flickr.FlickrAPI, opts *BackupPlanOptions, plan *BackupPlan) (*BackupPlan, error) {
	if !opts.All && len(opts.AlbumTitles) == 0 && len(opts.AlbumIDs) == 0 {
		return nil, fmt.Errorf("specify --all, --album, or --album-id")
	}

	// Collect matching albums from all pages of the album list.
	type matchedAlbum struct {
		id    string
		title string
	}
	var albums []matchedAlbum

	page := 1
	for {
		params := map[string]string{
			"user_id":  "me",
			"per_page": "500",
			"page":     fmt.Sprintf("%d", page),
		}

		var result flickr.PhotosetListResponse
		if err := client.Call(ctx, "flickr.photosets.getList", params, &result); err != nil {
			return nil, fmt.Errorf("listing albums: %w", err)
		}

		for _, ps := range result.Photosets.Photoset {
			if opts.All {
				albums = append(albums, matchedAlbum{id: ps.ID, title: ps.Title.Content})
				continue
			}

			for _, id := range opts.AlbumIDs {
				if ps.ID == id {
					albums = append(albums, matchedAlbum{id: ps.ID, title: ps.Title.Content})
					break
				}
			}

			for _, title := range opts.AlbumTitles {
				if matchAlbumTitle(ps.Title.Content, title) {
					albums = append(albums, matchedAlbum{id: ps.ID, title: ps.Title.Content})
					break
				}
			}
		}

		if result.Photosets.Pages == 0 || page >= result.Photosets.Pages {
			break
		}
		page++
	}

	for _, album := range albums {
		photos, err := getAlbumPhotos(ctx, client, album.id, album.title)
		if err != nil {
			return nil, fmt.Errorf("listing photos in album %q: %w", album.title, err)
		}
		plan.Items = append(plan.Items, photos...)
	}

	return plan, nil
}

// getAlbumPhotos returns all photos in the given album/photoset.
func getAlbumPhotos(ctx context.Context, client flickr.FlickrAPI, albumID, albumName string) ([]BackupItem, error) {
	var items []BackupItem
	page := 1
	for {
		params := map[string]string{
			"photoset_id": albumID,
			"per_page":    "500",
			"page":        fmt.Sprintf("%d", page),
			"extras":      flickr.DefaultExtras,
		}

		var result flickr.PhotosetGetPhotosResponse
		if err := client.Call(ctx, "flickr.photosets.getPhotos", params, &result); err != nil {
			return nil, err
		}

		for i := range result.Photoset.Photo {
			p := &result.Photoset.Photo[i]
			items = append(items, BackupItem{
				PhotoID:        p.ID,
				Title:          p.Title,
				AlbumID:        albumID,
				AlbumName:      albumName,
				Media:          p.Media,
				OriginalFormat: p.OriginalFormat,
				URLO:           p.URLO,
				URLK:           p.URLK,
				Secret:         p.Secret,
			})
		}

		if result.Photoset.Pages == 0 || page >= result.Photoset.Pages {
			break
		}
		page++
	}
	return items, nil
}

func buildUserPlan(ctx context.Context, client flickr.FlickrAPI, opts *BackupPlanOptions, plan *BackupPlan) (*BackupPlan, error) {
	userID := opts.UserID
	if userID == "" {
		userID = "me"
	}

	page := 1
	for {
		params := map[string]string{
			"user_id":  userID,
			"per_page": "500",
			"page":     fmt.Sprintf("%d", page),
			"extras":   flickr.DefaultExtras,
		}

		var result flickr.PhotoListResponse

		if err := client.Call(ctx, "flickr.people.getPhotos", params, &result); err != nil {
			return nil, fmt.Errorf("listing photos: %w", err)
		}

		for i := range result.Photos.Photo {
			p := &result.Photos.Photo[i]
			plan.Items = append(plan.Items, BackupItem{
				PhotoID:        p.ID,
				Title:          p.Title,
				Media:          p.Media,
				OriginalFormat: p.OriginalFormat,
				URLO:           p.URLO,
				URLK:           p.URLK,
				Secret:         p.Secret,
			})
		}

		if result.Photos.Pages == 0 || page >= result.Photos.Pages {
			break
		}
		page++
	}

	return plan, nil
}

func buildIDDirsPlan(ctx context.Context, client flickr.FlickrAPI, opts *BackupPlanOptions, plan *BackupPlan) (*BackupPlan, error) {
	return buildUserPlan(ctx, client, opts, plan)
}

func matchAlbumTitle(albumTitle, pattern string) bool {
	if strings.EqualFold(albumTitle, pattern) {
		return true
	}
	matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(albumTitle))
	return matched
}
