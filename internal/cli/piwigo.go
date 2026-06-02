package cli

import (
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

		// Read flags
		piwigoURL, _ := cmd.Flags().GetString("url")
		piwigoUser, _ := cmd.Flags().GetString("user")
		piwigoPassword, _ := cmd.Flags().GetString("password")
		albumPrefix, _ := cmd.Flags().GetString("album-prefix")
		importAlbum, _ := cmd.Flags().GetString("import-album")
		dedupe, _ := cmd.Flags().GetString("dedupe")
		hash, _ := cmd.Flags().GetString("hash")
		limit, _ := cmd.Flags().GetInt("limit")
		resume, _ := cmd.Flags().GetBool("resume")

		// Validate required flags
		if piwigoURL == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "--url is required"))
		}
		if piwigoUser == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "--user is required"))
		}
		if piwigoPassword == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "--password is required"))
		}

		// Read-only check
		if app.ReadOnly {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrReadOnlyViolation,
				"Operation blocked by --read-only flag",
				map[string]any{"command": "piwigo.import", "flag": "--read-only"},
			))
		}

		// Get Flickr client
		flickrClient, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, flickrClient); err != nil {
			return err
		}

		// Build import options
		opts := piwigo.ImportOptions{
			URL:         piwigoURL,
			Username:    piwigoUser,
			Password:    piwigoPassword,
			AlbumPrefix: albumPrefix,
			ImportAlbum: importAlbum,
			Dedupe:      dedupe,
			Hash:        hash,
			Limit:       limit,
			Resume:      resume,
		}

		// Create importer
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
			Flickr:    flickrClient,
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
	piwigoImportCmd.Flags().String("url", "", "Piwigo instance URL (required)")
	piwigoImportCmd.Flags().String("user", "", "Piwigo username (required)")
	piwigoImportCmd.Flags().String("password", "", "Piwigo password (required)")
	piwigoImportCmd.Flags().String("album-prefix", "", "prefix for created albums")
	piwigoImportCmd.Flags().String("import-album", "Imported from Piwigo", "import album name")
	piwigoImportCmd.Flags().String("dedupe", "checksum", "deduplication: checksum|none")
	piwigoImportCmd.Flags().String("hash", "md5", "hash algorithm: md5|sha1")
	piwigoImportCmd.Flags().Int("limit", 0, "limit number of imports (0 for all)")
	piwigoImportCmd.Flags().Bool("resume", false, "resume interrupted import")

	piwigoCmd.AddCommand(piwigoImportCmd)
}
