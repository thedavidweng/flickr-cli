package cli

import (
	"net/http"
	"path/filepath"

	"github.com/thedavidweng/flickr-cli/internal/backup"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup Flickr photos",
}

var backupAlbumsCmd = &cobra.Command{
	Use:   "albums",
	Short: "Backup photos by album",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "backup.albums",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		dest, _ := cmd.Flags().GetString("dest")
		albumTitles, _ := cmd.Flags().GetStringSlice("album")
		albumIDs, _ := cmd.Flags().GetStringSlice("album-id")
		allAlbums, _ := cmd.Flags().GetBool("all")
		size, _ := cmd.Flags().GetString("size")
		metadataFmt, _ := cmd.Flags().GetString("metadata")
		tmpl, _ := cmd.Flags().GetString("template")
		force, _ := cmd.Flags().GetBool("force")
		includeComments, _ := cmd.Flags().GetBool("include-comments")
		includeGeo, _ := cmd.Flags().GetBool("include-geo")
		includePools, _ := cmd.Flags().GetBool("include-pools")
		includeAlbums, _ := cmd.Flags().GetBool("include-albums")
		resume, _ := cmd.Flags().GetBool("resume")

		opts := backup.BackupPlanOptions{
			Mode:            backup.BackupAlbums,
			Dest:            dest,
			AlbumTitles:     albumTitles,
			AlbumIDs:        albumIDs,
			All:             allAlbums,
			Size:            size,
			Metadata:        metadataFmt,
			Force:           force,
			Resume:          resume,
			IncludeAlbums:   includeAlbums,
			IncludePools:    includePools,
			IncludeGeo:      includeGeo,
			IncludeComments: includeComments,
		}

		plan, err := backup.BuildPlan(cmd.Context(), client, opts)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		if app.DryRun {
			r.Human("Dry run: %d items planned for backup\n", len(plan.Items))
			return r.Success(meta, map[string]any{
				"planned": true,
				"items":   plan.Items,
				"count":   len(plan.Items),
			}, plan.Warnings)
		}

		dlItems := backupItemsToDownloadItems(plan.Items, dest, tmpl, metadataFmt, size)

		dl := &backup.Downloader{
			HTTP:        http.DefaultClient,
			Client:      client,
			Concurrency: app.Concurrency,
			Events:      output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
		}

		summary, err := dl.Download(cmd.Context(), dlItems, backup.DownloadOptions{
			Force:    force,
			Size:     size,
			Metadata: metadataFmt,
		})
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, summary, plan.Warnings)
		}

		r.Human("Backup complete: %d total, %d completed, %d skipped, %d failed\n",
			summary.Total, summary.Completed, summary.Skipped, summary.Failed)
		return nil
	},
}

var backupUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Backup photos by user with date/privacy filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "backup.user",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		dest, _ := cmd.Flags().GetString("dest")
		userID, _ := cmd.Flags().GetString("user-id")
		size, _ := cmd.Flags().GetString("size")
		metadataFmt, _ := cmd.Flags().GetString("metadata")
		tmpl, _ := cmd.Flags().GetString("template")
		force, _ := cmd.Flags().GetBool("force")
		resume, _ := cmd.Flags().GetBool("resume")

		opts := backup.BackupPlanOptions{
			Mode:     backup.BackupUser,
			Dest:     dest,
			UserID:   userID,
			Size:     size,
			Metadata: metadataFmt,
			Force:    force,
			Resume:   resume,
		}

		plan, err := backup.BuildPlan(cmd.Context(), client, opts)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		if app.DryRun {
			r.Human("Dry run: %d items planned for backup\n", len(plan.Items))
			return r.Success(meta, map[string]any{
				"planned": true,
				"items":   plan.Items,
				"count":   len(plan.Items),
			}, plan.Warnings)
		}

		dlItems := backupItemsToDownloadItems(plan.Items, dest, tmpl, metadataFmt, size)

		dl := &backup.Downloader{
			HTTP:        http.DefaultClient,
			Client:      client,
			Concurrency: app.Concurrency,
			Events:      output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
		}

		summary, err := dl.Download(cmd.Context(), dlItems, backup.DownloadOptions{
			Force:    force,
			Size:     size,
			Metadata: metadataFmt,
		})
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, summary, plan.Warnings)
		}

		r.Human("Backup complete: %d total, %d completed, %d skipped, %d failed\n",
			summary.Total, summary.Completed, summary.Skipped, summary.Failed)
		return nil
	},
}

var backupIDDirsCmd = &cobra.Command{
	Use:   "id-dirs",
	Short: "Stable full backup with ID-based directory structure",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "backup.id-dirs",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		dest, _ := cmd.Flags().GetString("dest")
		metadataFmt, _ := cmd.Flags().GetString("metadata")
		includeNotInAlbum, _ := cmd.Flags().GetBool("include-not-in-album")
		includeAlbums, _ := cmd.Flags().GetBool("include-albums")
		includePools, _ := cmd.Flags().GetBool("include-pools")
		includeGeo, _ := cmd.Flags().GetBool("include-geo")
		includeComments, _ := cmd.Flags().GetBool("include-comments")
		force, _ := cmd.Flags().GetBool("force")
		resume, _ := cmd.Flags().GetBool("resume")

		opts := backup.BackupPlanOptions{
			Mode:              backup.BackupIDDirs,
			Dest:              dest,
			Size:              "original",
			Metadata:          metadataFmt,
			Force:             force,
			Resume:            resume,
			IncludeNotInAlbum: includeNotInAlbum,
			IncludeAlbums:     includeAlbums,
			IncludePools:      includePools,
			IncludeGeo:        includeGeo,
			IncludeComments:   includeComments,
		}

		plan, err := backup.BuildPlan(cmd.Context(), client, opts)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		if app.DryRun {
			r.Human("Dry run: %d items planned for backup\n", len(plan.Items))
			return r.Success(meta, map[string]any{
				"planned": true,
				"items":   plan.Items,
				"count":   len(plan.Items),
			}, plan.Warnings)
		}

		dlItems := backupIDDirsToDownloadItems(plan.Items, dest, metadataFmt)

		dl := &backup.Downloader{
			HTTP:        http.DefaultClient,
			Client:      client,
			Concurrency: app.Concurrency,
			Events:      output.EventWriter{Enabled: app.Events, Err: cmd.ErrOrStderr()},
		}

		summary, err := dl.Download(cmd.Context(), dlItems, backup.DownloadOptions{
			Force:    force,
			Size:     "original",
			Metadata: metadataFmt,
		})
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, summary, plan.Warnings)
		}

		r.Human("Backup complete: %d total, %d completed, %d skipped, %d failed\n",
			summary.Total, summary.Completed, summary.Skipped, summary.Failed)
		return nil
	},
}

// backupItemsToDownloadItems converts plan items to download items using a
// standard directory layout.
func backupItemsToDownloadItems(items []backup.BackupItem, dest, tmpl, metadataFmt, size string) []backup.DownloadItem {
	dlItems := make([]backup.DownloadItem, 0, len(items))
	for _, item := range items {
		safeTitle := backup.SafeName(item.Title, item.PhotoID)
		var photoDir string
		if tmpl == "archive" {
			photoDir = filepath.Join(dest, safeTitle)
		} else {
			photoDir = dest
		}

		dl := backup.DownloadItem{
			PhotoID:   item.PhotoID,
			FilePath:  filepath.Join(photoDir, item.PhotoID+".jpg"),
			SizeLabel: size,
		}

		switch metadataFmt {
		case "json":
			dl.MetadataPathJSON = filepath.Join(photoDir, item.PhotoID+".json")
		case "yaml":
			dl.MetadataPathYAML = filepath.Join(photoDir, item.PhotoID+".yaml")
		case "both":
			dl.MetadataPathJSON = filepath.Join(photoDir, item.PhotoID+".json")
			dl.MetadataPathYAML = filepath.Join(photoDir, item.PhotoID+".yaml")
		}

		dlItems = append(dlItems, dl)
	}
	return dlItems
}

// backupIDDirsToDownloadItems converts plan items to download items using
// the stable ID-based directory layout (hash/hash/id/id.ext).
func backupIDDirsToDownloadItems(items []backup.BackupItem, dest, metadataFmt string) []backup.DownloadItem {
	dlItems := make([]backup.DownloadItem, 0, len(items))
	for _, item := range items {
		dl := backup.DownloadItem{
			PhotoID:   item.PhotoID,
			FilePath:  backup.IDDirsPath(dest, item.PhotoID, "jpg"),
			SizeLabel: "original",
		}

		switch metadataFmt {
		case "json":
			dl.MetadataPathJSON = backup.IDDirsPath(dest, item.PhotoID, "json")
		case "yaml":
			dl.MetadataPathYAML = backup.IDDirsPath(dest, item.PhotoID, "yaml")
		case "both":
			dl.MetadataPathJSON = backup.IDDirsPath(dest, item.PhotoID, "json")
			dl.MetadataPathYAML = backup.IDDirsPath(dest, item.PhotoID, "yaml")
		}

		dlItems = append(dlItems, dl)
	}
	return dlItems
}

func init() {
	backupAlbumsCmd.Flags().String("dest", "./flickr-backup", "destination directory")
	backupAlbumsCmd.Flags().StringSlice("album", nil, "album title or glob (repeatable)")
	backupAlbumsCmd.Flags().StringSlice("album-id", nil, "album ID (repeatable)")
	backupAlbumsCmd.Flags().Bool("all", false, "include all albums")
	backupAlbumsCmd.Flags().String("size", "original", "size: original|large|medium")
	backupAlbumsCmd.Flags().String("metadata", "json", "metadata format: json|yaml|both")
	backupAlbumsCmd.Flags().String("template", "archive", "template: archive|<dir>")
	backupAlbumsCmd.Flags().Bool("force", false, "overwrite existing files")
	backupAlbumsCmd.Flags().Bool("include-comments", false, "include comments")
	backupAlbumsCmd.Flags().Bool("include-geo", false, "include geo data")
	backupAlbumsCmd.Flags().Bool("include-pools", false, "include pools")
	backupAlbumsCmd.Flags().Bool("include-albums", false, "include album memberships")
	backupAlbumsCmd.Flags().Bool("resume", false, "resume interrupted backup")

	backupUserCmd.Flags().String("dest", "./flickr-backup", "destination directory")
	backupUserCmd.Flags().String("user-id", "me", "user ID or 'me'")
	backupUserCmd.Flags().String("min-upload-date", "", "minimum upload date")
	backupUserCmd.Flags().String("max-upload-date", "", "maximum upload date")
	backupUserCmd.Flags().String("min-taken-date", "", "minimum taken date")
	backupUserCmd.Flags().String("max-taken-date", "", "maximum taken date")
	backupUserCmd.Flags().String("privacy", "", "privacy level filter")
	backupUserCmd.Flags().String("album-id", "", "filter by album ID")
	backupUserCmd.Flags().String("size", "original", "size: original|large|medium")
	backupUserCmd.Flags().String("metadata", "json", "metadata format: json|yaml|both")
	backupUserCmd.Flags().String("template", "archive", "template: archive|<dir>")
	backupUserCmd.Flags().Bool("resume", false, "resume interrupted backup")

	backupIDDirsCmd.Flags().String("dest", "./flickr-backup", "destination directory")
	backupIDDirsCmd.Flags().String("metadata", "both", "metadata format: json|yaml|both")
	backupIDDirsCmd.Flags().Bool("include-not-in-album", true, "include photos not in any album")
	backupIDDirsCmd.Flags().Bool("include-albums", true, "include album memberships")
	backupIDDirsCmd.Flags().Bool("include-pools", true, "include pool memberships")
	backupIDDirsCmd.Flags().Bool("include-geo", true, "include geo data")
	backupIDDirsCmd.Flags().Bool("include-comments", false, "include comments")
	backupIDDirsCmd.Flags().Bool("force", false, "overwrite existing files")
	backupIDDirsCmd.Flags().Bool("resume", true, "resume interrupted backup")

	backupCmd.AddCommand(backupAlbumsCmd)
	backupCmd.AddCommand(backupUserCmd)
	backupCmd.AddCommand(backupIDDirsCmd)
}
