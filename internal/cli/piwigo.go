package cli

import (
	"os"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/piwigo"
	"github.com/thedavidweng/flickr-cli/internal/safety"
	"github.com/spf13/cobra"
)

var piwigoCmd = &cobra.Command{
	Use:   "piwigo",
	Short: "Piwigo migration tools",
}

var piwigoImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import photos from Piwigo to Flickr",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "piwigo.import",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		// Read all MySQL flags
		uploads, _ := cmd.Flags().GetString("uploads")
		mysqlDB, _ := cmd.Flags().GetString("mysql-db")
		mysqlHost, _ := cmd.Flags().GetString("mysql-host")
		mysqlPort, _ := cmd.Flags().GetInt("mysql-port")
		mysqlUser, _ := cmd.Flags().GetString("mysql-user")
		mysqlPassword, _ := cmd.Flags().GetString("mysql-password")
		mysqlPasswordEnv, _ := cmd.Flags().GetString("mysql-password-env")
		tablePrefix, _ := cmd.Flags().GetString("table-prefix")
		albumPrefix, _ := cmd.Flags().GetString("album-prefix")
		importAlbum, _ := cmd.Flags().GetString("import-album")
		dedupe, _ := cmd.Flags().GetString("dedupe")
		hash, _ := cmd.Flags().GetString("hash")
		limit, _ := cmd.Flags().GetInt("limit")
		resume, _ := cmd.Flags().GetBool("resume")

		// Validate required flags
		if uploads == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "--uploads is required"))
		}
		if mysqlDB == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "--mysql-db is required"))
		}

		// Resolve password from environment variable if specified
		password := mysqlPassword
		if mysqlPasswordEnv != "" {
			password = os.Getenv(mysqlPasswordEnv)
			if password == "" {
				return r.Failure(meta, output.Errorf(model.ErrConfig, "environment variable %q is not set", mysqlPasswordEnv))
			}
		}

		// Read-only check: block all write operations
		if app.ReadOnly {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrReadOnlyViolation,
				"Operation blocked by --read-only flag",
				map[string]any{"command": "piwigo.import", "flag": "--read-only"},
			))
		}

		// Build DB config
		dbConfig := piwigo.DBConfig{
			Host:        mysqlHost,
			Port:        mysqlPort,
			DB:          mysqlDB,
			User:        mysqlUser,
			Password:    password,
			TablePrefix: tablePrefix,
		}

		// Open MySQL connection (skip for dry-run since no actual work is needed)
		if !app.DryRun {
			db, err := piwigo.Open(cmd.Context(), dbConfig)
			if err != nil {
				return r.Failure(meta, output.Errorf(model.ErrConfig, "opening database: %v", err))
			}
			defer db.Close()
		}

		// Build import options
		opts := piwigo.ImportOptions{
			Uploads:     uploads,
			DB:          dbConfig,
			AlbumPrefix: albumPrefix,
			ImportAlbum: importAlbum,
			Dedupe:      dedupe,
			Hash:        hash,
			Limit:       limit,
			Resume:      resume,
		}

		// Create importer with safety gate
		imp := &piwigo.Importer{
			Events: output.EventWriter{
				Enabled: app.Events,
				Err:     cmd.ErrOrStderr(),
			},
			Gate: safety.GateInput{
				ReadOnly: app.ReadOnly,
				DryRun:   app.DryRun,
				Confirm:  app.Confirm,
			},
			Profile:   app.Profile,
			RequestID: app.RequestID,
		}

		// Run import
		summary, err := imp.Import(cmd.Context(), opts)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.DryRun {
			r.Human("Dry run: %d planned imports\n", summary.Planned)
			return r.Success(meta, summary, nil)
		}

		r.Human("Import complete: %d succeeded, %d skipped, %d failed\n", summary.Succeeded, summary.Skipped, summary.Failed)
		return r.Success(meta, summary, nil)
	},
}

func init() {
	piwigoImportCmd.Flags().String("uploads", "", "Piwigo uploads root directory (required)")
	piwigoImportCmd.Flags().String("mysql-host", "localhost", "MySQL host")
	piwigoImportCmd.Flags().Int("mysql-port", 3306, "MySQL port")
	piwigoImportCmd.Flags().String("mysql-db", "", "MySQL database name")
	piwigoImportCmd.Flags().String("mysql-user", "", "MySQL user")
	piwigoImportCmd.Flags().String("mysql-password", "", "MySQL password")
	piwigoImportCmd.Flags().String("mysql-password-env", "", "env var name containing MySQL password")
	piwigoImportCmd.Flags().String("table-prefix", "", "Piwigo table prefix")
	piwigoImportCmd.Flags().String("album-prefix", "", "prefix for created albums")
	piwigoImportCmd.Flags().String("import-album", "Imported from Piwigo", "import album name")
	piwigoImportCmd.Flags().String("dedupe", "checksum", "deduplication: checksum|none")
	piwigoImportCmd.Flags().String("hash", "md5", "hash algorithm: md5|sha1")
	piwigoImportCmd.Flags().Int("limit", 0, "limit number of imports (0 for all)")
	piwigoImportCmd.Flags().Bool("resume", false, "resume interrupted import")

	piwigoCmd.AddCommand(piwigoImportCmd)
}
