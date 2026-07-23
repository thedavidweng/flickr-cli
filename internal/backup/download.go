package backup

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// DownloadConfig holds the runtime configuration for a download operation.
// It is the single set of options the CLI needs to provide — all path
// construction, downloader wiring, and concurrency happen behind this.
type DownloadConfig struct {
	Dest        string
	Size        string
	SizeMax     int
	Metadata    string
	Force       bool
	Exif        bool
	Concurrency int
	Events      *output.EventWriter
}

// DownloadByPlan builds a plan from opts, constructs download items with
// layout-appropriate paths, and runs the downloader. This is the deep
// entry point for backup-mode downloads (album, all, id-dirs, user).
func DownloadByPlan(ctx context.Context, client flickr.FlickrAPI, httpClient *http.Client, planOpts *BackupPlanOptions, cfg DownloadConfig) (*DownloadSummary, error) {
	plan, err := BuildPlan(ctx, client, planOpts)
	if err != nil {
		return nil, fmt.Errorf("building plan: %w", err)
	}

	if len(plan.Items) == 0 {
		return &DownloadSummary{Total: 0}, nil
	}

	items := planToItems(plan, planOpts)
	return runDownload(ctx, client, httpClient, items, cfg)
}

// DownloadByIDs constructs download items from explicit photo IDs and runs
// the downloader. This is the deep entry point for direct-ID downloads.
func DownloadByIDs(ctx context.Context, client flickr.FlickrAPI, httpClient *http.Client, photoIDs []string, cfg DownloadConfig) (*DownloadSummary, error) {
	items := idsToItems(photoIDs, cfg)
	return runDownload(ctx, client, httpClient, items, cfg)
}

// PlanModeFromFlags converts CLI layout/album flags to a PlanMode.
func PlanModeFromFlags(layout string, all, hasAlbums bool) PlanMode {
	switch layout {
	case "id-dirs":
		return BackupIDDirs
	case "album":
		return BackupAlbums
	default:
		if all || hasAlbums {
			return BackupAlbums
		}
		return BackupUser
	}
}

// planToItems converts plan items to download items with layout-appropriate paths.
func planToItems(plan *BackupPlan, opts *BackupPlanOptions) []DownloadItem {
	items := make([]DownloadItem, len(plan.Items))
	for i := range plan.Items {
		item := &plan.Items[i]
		ext := flickr.DeriveExtension("", item.Media, item.OriginalFormat)
		var filePath string

		switch opts.Mode {
		case BackupIDDirs:
			filePath = IDDirsPath(opts.Dest, item.PhotoID, ext)
		default:
			fileName := SafeName(item.Title, item.PhotoID) + "." + ext
			if item.AlbumName != "" {
				filePath = filepath.Join(opts.Dest, SafeName(item.AlbumName, item.AlbumID), fileName)
			} else {
				filePath = filepath.Join(opts.Dest, fileName)
			}
		}

		dlItem := DownloadItem{
			PhotoID:        item.PhotoID,
			FilePath:       filePath,
			SizeLabel:      opts.Size,
			Media:          item.Media,
			OriginalFormat: item.OriginalFormat,
		}
		if opts.Metadata == "json" || opts.Metadata == "both" {
			dlItem.MetadataPathJSON = filePath + ".json"
		}
		if opts.Metadata == "yaml" || opts.Metadata == "both" {
			dlItem.MetadataPathYAML = filePath + ".yaml"
		}
		items[i] = dlItem
	}
	return items
}

// idsToItems constructs download items from explicit photo IDs.
func idsToItems(photoIDs []string, cfg DownloadConfig) []DownloadItem {
	items := make([]DownloadItem, len(photoIDs))
	for i, photoID := range photoIDs {
		filePath := filepath.Join(cfg.Dest, photoID+".jpg")
		item := DownloadItem{
			PhotoID:   photoID,
			FilePath:  filePath,
			SizeLabel: cfg.Size,
		}
		if cfg.Metadata == "json" || cfg.Metadata == "both" {
			item.MetadataPathJSON = filePath + ".json"
		}
		if cfg.Metadata == "yaml" || cfg.Metadata == "both" {
			item.MetadataPathYAML = filePath + ".yaml"
		}
		items[i] = item
	}
	return items
}

// runDownload wires a Downloader and executes the download.
func runDownload(ctx context.Context, client flickr.FlickrAPI, httpClient *http.Client, items []DownloadItem, cfg DownloadConfig) (*DownloadSummary, error) {
	dl := &Downloader{
		HTTP:        httpClient,
		Client:      client,
		Concurrency: cfg.Concurrency,
		Events:      cfg.Events,
	}

	return dl.Download(ctx, items, DownloadOptions{
		Force:    cfg.Force,
		Size:     cfg.Size,
		SizeMax:  cfg.SizeMax,
		Exif:     cfg.Exif,
		Metadata: cfg.Metadata,
	})
}

// DefaultHTTPClient returns an HTTP client suitable for photo downloads.
func DefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 120 * time.Second}
}
