package cli

import (
	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/backup"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
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

		dlCfg := &backup.DownloadConfig{
			Dest:        dest,
			Size:        size,
			SizeMax:     sizeMax,
			Metadata:    metadata,
			Force:       force,
			Exif:        exif,
			Concurrency: ctx.App.Concurrency,
			Events:      &output.EventWriter{Enabled: ctx.App.Events, Err: ctx.Cmd.ErrOrStderr()},
		}

		if backupMode {
			planOpts := &backup.BackupPlanOptions{
				Mode:        backup.PlanModeFromFlags(layout, allAlbums, len(albumTitles) > 0 || len(albumIDs) > 0),
				Dest:        dest,
				AlbumTitles: albumTitles,
				AlbumIDs:    albumIDs,
				All:         allAlbums,
				Size:        size,
				SizeMax:     sizeMax,
				Metadata:    metadata,
				Force:       force,
				Exif:        exif,
			}

			// Dry-run preview needs the plan item count.
			if ctx.App.DryRun {
				plan, err := backup.BuildPlan(ctx.Cmd.Context(), ctx.Client, planOpts)
				if err != nil {
					return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "%v", err))
				}
				ctx.R.Human("Would download %d photos to %s\n", len(plan.Items), dest)
				return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "total": len(plan.Items)}, nil)
			}

			summary, err := backup.DownloadByPlan(ctx.Cmd.Context(), ctx.Client, backup.DefaultHTTPClient(), planOpts, dlCfg)
			if err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			if summary.Total == 0 {
				ctx.R.Human("No photos to download\n")
			}
			if ctx.App.JSON {
				return ctx.R.Success(ctx.Meta, map[string]any{"summary": summary, "dest": dest}, nil)
			}
			ctx.R.Human("\nSummary: %d total, %d completed, %d skipped, %d failed\n",
				summary.Total, summary.Completed, summary.Skipped, summary.Failed)
			return nil
		}

		if len(ctx.Args) == 0 {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "specify photo IDs, --album, --album-id, or --all"))
		}

		if ctx.App.DryRun {
			planned := make([]map[string]any, 0, len(ctx.Args))
			for _, id := range ctx.Args {
				planned = append(planned, map[string]any{"photo_id": id, "dest": dest, "size": size})
			}
			ctx.R.Human("Would download %d photos to %s\n", len(ctx.Args), dest)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "items": planned}, nil)
		}

		summary, err := backup.DownloadByIDs(ctx.Cmd.Context(), ctx.Client, backup.DefaultHTTPClient(), ctx.Args, dlCfg)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"summary": summary,
				"dest":    dest,
				"size":    size,
			}, nil)
		}

		ctx.R.Human("\nSummary: %d total, %d completed, %d skipped, %d failed\n",
			summary.Total, summary.Completed, summary.Skipped, summary.Failed)
		return nil
	}),
}
