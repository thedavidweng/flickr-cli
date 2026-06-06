package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/cache"
	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local cache",
}

var cacheSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync albums and photos to local cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "cache.sync",
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

		albums, _ := cmd.Flags().GetBool("albums")
		photos, _ := cmd.Flags().GetBool("photos")

		dbPath := config.DefaultCachePath(app.Profile)
		db, err := cache.Open(dbPath, app.Profile)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrCache, "opening cache: %v", err))
		}
		defer db.Close()

		result, err := cache.Sync(cmd.Context(), db, client, cache.SyncOptions{
			Albums: albums,
			Photos: photos,
		})
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrCache, "syncing cache: %v", err))
		}

		if app.JSON {
			return r.Success(meta, result, nil)
		}

		r.Human("Synced %d albums and %d photos\n", result.AlbumsSynced, result.PhotosSynced)
		return nil
	},
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "cache.stats",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		dbPath := config.DefaultCachePath(app.Profile)
		db, err := cache.Open(dbPath, app.Profile)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrCache, "opening cache: %v", err))
		}
		defer db.Close()

		stats, err := db.Stats()
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrCache, "getting stats: %v", err))
		}

		fileSize, _ := cache.StatFile(dbPath)
		stats.Path = dbPath

		if app.JSON {
			return r.Success(meta, map[string]any{
				"path":       dbPath,
				"counts":     stats.Counts,
				"file_bytes": fileSize,
			}, nil)
		}

		r.Human("Cache path: %s\n", dbPath)
		r.Human("File size:  %d bytes\n", fileSize)
		r.Human("Albums:     %d\n", stats.Counts.Albums)
		r.Human("Photos:     %d\n", stats.Counts.Photos)
		r.Human("Checksums:  %d\n", stats.Counts.Checksums)
		r.Human("Jobs:       %d\n", stats.Counts.Jobs)
		return nil
	},
}

var cacheCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove expired cache entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "cache.cleanup",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		olderThan, _ := cmd.Flags().GetDuration("older-than")

		dbPath := config.DefaultCachePath(app.Profile)
		db, err := cache.Open(dbPath, app.Profile)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrCache, "opening cache: %v", err))
		}
		defer db.Close()

		count, err := db.Cleanup(olderThan)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrCache, "cleanup: %v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{"removed": count}, nil)
		}

		r.Human("Removed %d entries older than %s\n", count, olderThan)
		return nil
	},
}

func init() {
	cacheSyncCmd.Flags().Bool("albums", true, "sync albums")
	cacheSyncCmd.Flags().Bool("photos", false, "sync photos")
	cacheCleanupCmd.Flags().Duration("older-than", 720*time.Hour, "remove entries older than")

	cacheCmd.AddCommand(cacheSyncCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheCleanupCmd)
}
