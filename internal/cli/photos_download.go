package cli

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/thedavidweng/flickr-cli/internal/backup"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var photosDownloadCmd = &cobra.Command{
	Use:   "download [photo-id...]",
	Short: "Download photos",
	Long: `Download photos from Flickr. Supports three modes:

1. By photo ID: flickr photos download 12345 67890
2. By album: flickr photos download --album "Vacation" --dest ./backup
3. All photos: flickr photos download --all --dest ./backup --layout id-dirs`,
	Args: cobra.ArbitraryArgs,
	RunE: withAuth("photos.download", func(ctx *CmdContext) error {
		dest, _ := ctx.Cmd.Flags().GetString("dest")
		size, _ := ctx.Cmd.Flags().GetString("size")
		sizeMax, _ := ctx.Cmd.Flags().GetInt("size-max")
		metadata, _ := ctx.Cmd.Flags().GetString("metadata")
		force, _ := ctx.Cmd.Flags().GetBool("force")
		layout, _ := ctx.Cmd.Flags().GetString("layout")
		albumTitles, _ := ctx.Cmd.Flags().GetStringSlice("album")
		albumIDs, _ := ctx.Cmd.Flags().GetStringSlice("album-id")
		allAlbums, _ := ctx.Cmd.Flags().GetBool("all")
		exif, _ := ctx.Cmd.Flags().GetBool("exif")

		if dest == "" {
			dest = "./flickr-backup"
		}

		backupMode := allAlbums || len(albumTitles) > 0 || len(albumIDs) > 0 || layout != ""

		if backupMode {
			return downloadViaBackup(ctx.Cmd, ctx.Client, ctx.R, ctx.Meta, ctx.App, backup.BackupPlanOptions{
				Mode:        backupModeToPlanMode(layout, allAlbums, len(albumTitles) > 0 || len(albumIDs) > 0),
				Dest:        dest,
				AlbumTitles: albumTitles,
				AlbumIDs:    albumIDs,
				All:         allAlbums,
				Size:        size,
				SizeMax:     sizeMax,
				Metadata:    metadata,
				Force:       force,
				Exif:        exif,
			})
		}

		if len(ctx.Args) == 0 {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "specify photo IDs, --album, --album-id, or --all"))
		}

		return downloadByIDs(ctx.Cmd, ctx.Client, ctx.R, ctx.Meta, ctx.App, ctx.Args, dest, size, sizeMax, metadata, force, exif)
	}),
}

// backupModeToPlanMode converts CLI flags to backup plan mode.
func backupModeToPlanMode(layout string, all bool, hasAlbums bool) backup.PlanMode {
	switch layout {
	case "id-dirs":
		return backup.BackupIDDirs
	case "album":
		return backup.BackupAlbums
	default:
		if all || hasAlbums {
			return backup.BackupAlbums
		}
		return backup.BackupUser
	}
}

// downloadViaBackup handles backup mode using the backup package.
func downloadViaBackup(cmd *cobra.Command, client *flickr.Client, r output.Renderer, meta output.RuntimeMetaInput, app *AppContext, opts backup.BackupPlanOptions) error {
	plan, err := backup.BuildPlan(cmd.Context(), client, opts)
	if err != nil {
		return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
	}

	if len(plan.Items) == 0 {
		r.Human("No photos to download\n")
		return r.Success(meta, map[string]any{"total": 0}, nil)
	}

	if app.DryRun {
		r.Human("Would download %d photos to %s\n", len(plan.Items), opts.Dest)
		return r.Success(meta, map[string]any{"planned": true, "total": len(plan.Items)}, nil)
	}

	items := make([]backup.DownloadItem, len(plan.Items))
	for i, item := range plan.Items {
		// Use a placeholder extension; the downloader will fix it after resolving the download URL.
		ext := flickr.DeriveExtension("", item.Media, item.OriginalFormat)
		var filePath string

		switch opts.Mode {
		case backup.BackupIDDirs:
			filePath = backup.IDDirsPath(opts.Dest, item.PhotoID, ext)
		default:
			fileName := backup.SafeName(item.Title, item.PhotoID) + "." + ext
			if item.AlbumName != "" {
				filePath = filepath.Join(opts.Dest, backup.SafeName(item.AlbumName, item.AlbumID), fileName)
			} else {
				filePath = filepath.Join(opts.Dest, fileName)
			}
		}

		dlItem := backup.DownloadItem{
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

	downloader := &backup.Downloader{
		HTTP:        &http.Client{Timeout: 120 * time.Second},
		Client:      client,
		Concurrency: app.Concurrency,
		Events:      &output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
	}

	summary, err := downloader.Download(cmd.Context(), items, backup.DownloadOptions{
		Force:    opts.Force,
		Size:     opts.Size,
		SizeMax:  opts.SizeMax,
		Exif:     opts.Exif,
	})
	if err != nil {
		return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
	}

	if app.JSON {
		return r.Success(meta, map[string]any{
			"summary": summary,
			"dest":    opts.Dest,
		}, nil)
	}

	r.Human("\nSummary: %d total, %d completed, %d skipped, %d failed\n",
		summary.Total, summary.Completed, summary.Skipped, summary.Failed)
	return nil
}

// downloadByIDs handles direct download of specific photo IDs using backup.Downloader.
func downloadByIDs(cmd *cobra.Command, client *flickr.Client, r output.Renderer, meta output.RuntimeMetaInput, app *AppContext, photoIDs []string, dest, size string, sizeMax int, metadata string, force, exif bool) error {
	if app.DryRun {
		var planned []map[string]any
		for _, id := range photoIDs {
			planned = append(planned, map[string]any{"photo_id": id, "dest": dest, "size": size})
		}
		r.Human("Would download %d photos to %s\n", len(photoIDs), dest)
		return r.Success(meta, map[string]any{"planned": true, "items": planned}, nil)
	}

	items := make([]backup.DownloadItem, len(photoIDs))
	for i, photoID := range photoIDs {
		// Use a placeholder extension; the downloader will fix it after resolving the download URL.
		filePath := filepath.Join(dest, photoID+".jpg")
		item := backup.DownloadItem{
			PhotoID:   photoID,
			FilePath:  filePath,
			SizeLabel: size,
		}
		if metadata == "json" || metadata == "both" {
			item.MetadataPathJSON = filePath + ".json"
		}
		if metadata == "yaml" || metadata == "both" {
			item.MetadataPathYAML = filePath + ".yaml"
		}
		items[i] = item
	}

	downloader := &backup.Downloader{
		HTTP:        &http.Client{Timeout: 120 * time.Second},
		Client:      client,
		Concurrency: app.Concurrency,
		Events:      &output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
	}

	dlSummary, err := downloader.Download(cmd.Context(), items, backup.DownloadOptions{
		Force:   force,
		Size:    size,
		SizeMax: sizeMax,
		Exif:    exif,
	})
	if err != nil {
		return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
	}

	if app.JSON {
		return r.Success(meta, map[string]any{
			"summary": dlSummary,
			"dest":    dest,
			"size":    size,
		}, nil)
	}

	r.Human("\nSummary: %d total, %d completed, %d skipped, %d failed\n",
		dlSummary.Total, dlSummary.Completed, dlSummary.Skipped, dlSummary.Failed)
	return nil
}
