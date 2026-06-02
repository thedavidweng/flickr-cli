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
	Mode              PlanMode
	Dest              string
	AlbumTitles       []string
	AlbumIDs          []string
	All               bool
	UserID            string
	Size              string
	Metadata          string
	Force             bool
	Resume            bool
	IncludeNotInAlbum bool
	IncludeAlbums     bool
	IncludePools      bool
	IncludeGeo        bool
	IncludeComments   bool
}

// BackupPlan contains the items to back up.
type BackupPlan struct {
	Items    []BackupItem `json:"items"`
	Warnings []string     `json:"warnings,omitempty"`
}

// BackupItem is a single photo to back up.
type BackupItem struct {
	PhotoID string `json:"photo_id"`
	Title   string `json:"title,omitempty"`
	AlbumID string `json:"album_id,omitempty"`
}

// BuildPlan creates a backup plan from the given options.
func BuildPlan(ctx context.Context, client *flickr.Client, opts BackupPlanOptions) (*BackupPlan, error) {
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

func buildAlbumPlan(ctx context.Context, client *flickr.Client, opts BackupPlanOptions, plan *BackupPlan) (*BackupPlan, error) {
	if !opts.All && len(opts.AlbumTitles) == 0 && len(opts.AlbumIDs) == 0 {
		return nil, fmt.Errorf("specify --all, --album, or --album-id")
	}

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
				Photos int `json:"photos"`
			} `json:"photoset"`
		} `json:"photosets"`
	}

	if err := client.Call(ctx, "flickr.photosets.getList", params, &result); err != nil {
		return nil, fmt.Errorf("listing albums: %w", err)
	}

	for _, ps := range result.Photosets.Photoset {
		if opts.All {
			plan.Items = append(plan.Items, BackupItem{
				PhotoID: ps.ID,
				Title:   ps.Title.Content,
				AlbumID: ps.ID,
			})
			continue
		}

		for _, id := range opts.AlbumIDs {
			if ps.ID == id {
				plan.Items = append(plan.Items, BackupItem{
					PhotoID: ps.ID,
					Title:   ps.Title.Content,
					AlbumID: ps.ID,
				})
				break
			}
		}

		for _, title := range opts.AlbumTitles {
			if matchAlbumTitle(ps.Title.Content, title) {
				plan.Items = append(plan.Items, BackupItem{
					PhotoID: ps.ID,
					Title:   ps.Title.Content,
					AlbumID: ps.ID,
				})
				break
			}
		}
	}

	return plan, nil
}

func buildUserPlan(ctx context.Context, client *flickr.Client, opts BackupPlanOptions, plan *BackupPlan) (*BackupPlan, error) {
	userID := opts.UserID
	if userID == "" {
		userID = "me"
	}

	params := map[string]string{
		"user_id":  userID,
		"per_page": "500",
	}

	var result struct {
		Photos struct {
			Photo []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"photo"`
		} `json:"photos"`
	}

	if err := client.Call(ctx, "flickr.people.getPhotos", params, &result); err != nil {
		return nil, fmt.Errorf("listing photos: %w", err)
	}

	for _, p := range result.Photos.Photo {
		plan.Items = append(plan.Items, BackupItem{
			PhotoID: p.ID,
			Title:   p.Title,
		})
	}

	return plan, nil
}

func buildIDDirsPlan(ctx context.Context, client *flickr.Client, opts BackupPlanOptions, plan *BackupPlan) (*BackupPlan, error) {
	return buildUserPlan(ctx, client, opts, plan)
}

func matchAlbumTitle(albumTitle, pattern string) bool {
	if strings.EqualFold(albumTitle, pattern) {
		return true
	}
	matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(albumTitle))
	return matched
}
