package cli

import (
	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/piwigo"
	"github.com/thedavidweng/flickr-cli/internal/safety"
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
		r := newRenderer(app, cmd)
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
		limit, _ := cmd.Flags().GetInt("limit")

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

		// Safety gate (consistent with all other mutation commands)
		gate := safety.Check(safety.GateInput{
			ReadOnly: app.ReadOnly,
			DryRun:   app.DryRun,
			Confirm:  app.Confirm,
		}, safety.Mutation{
			Command: "piwigo.import",
			Method:  "flickr.upload",
			Risk:    safety.ClassifyRisk("piwigo.import"),
		})
		if gate.Error != nil {
			return r.Failure(meta, *gate.Error)
		}
		if gate.Planned {
			r.Human("Dry run: piwigo import would connect to %s\n", piwigoURL)
			return r.Success(meta, map[string]any{"planned": true}, nil)
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
			Limit:       limit,
		}

		// Create importer
		imp := &piwigo.Importer{
			Events: &output.EventWriter{
				Enabled: app.Events,
				Err:     cmd.ErrOrStderr(),
			},
			Profile:   app.Profile,
			RequestID: app.RequestID,
			Flickr:    flickrClient,
		}

		// Run import
		summary, err := imp.Import(cmd.Context(), &opts)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
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
	piwigoImportCmd.Flags().Int("limit", 0, "limit number of imports (0 for all)")

	piwigoCmd.AddCommand(piwigoImportCmd)
}
