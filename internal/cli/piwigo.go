package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/piwigo"
)

var piwigoCmd = &cobra.Command{
	Use:   "piwigo",
	Short: "Piwigo migration tools",
}

var piwigoImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import photos from Piwigo to Flickr",
	RunE: withAuth("piwigo.import", func(ctx *CmdContext) error {
		piwigoURL, _ := ctx.Cmd.Flags().GetString("url")
		piwigoUser, _ := ctx.Cmd.Flags().GetString("user")
		piwigoPassword, _ := ctx.Cmd.Flags().GetString("password")
		albumPrefix, _ := ctx.Cmd.Flags().GetString("album-prefix")
		importAlbum, _ := ctx.Cmd.Flags().GetString("import-album")
		dedupe, _ := ctx.Cmd.Flags().GetString("dedupe")
		limit, _ := ctx.Cmd.Flags().GetInt("limit")

		if piwigoURL == "" {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "--url is required"))
		}
		if piwigoUser == "" {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "--user is required"))
		}
		if piwigoPassword == "" {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "--password is required"))
		}

		return ctx.runMutation(mutationSpec{
			Command:  "piwigo.import",
			Method:   "flickr.upload",
			Resource: map[string]any{"url": piwigoURL},
			PlanMsg:  fmt.Sprintf("Dry run: piwigo import would connect to %s\n", piwigoURL),
		}, func() (any, error) {
			opts := piwigo.ImportOptions{
				URL:         piwigoURL,
				Username:    piwigoUser,
				Password:    piwigoPassword,
				AlbumPrefix: albumPrefix,
				ImportAlbum: importAlbum,
				Dedupe:      dedupe,
				Limit:       limit,
			}
			imp := &piwigo.Importer{
				Events:    &output.EventWriter{Enabled: ctx.App.Events, Err: ctx.Cmd.ErrOrStderr()},
				Profile:   ctx.App.Profile,
				RequestID: ctx.App.RequestID,
				Flickr:    ctx.Client,
			}
			summary, err := imp.Import(ctx.Cmd.Context(), &opts)
			if err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Import complete: %d succeeded, %d skipped, %d failed\n", summary.Succeeded, summary.Skipped, summary.Failed)
			return summary, nil
		})
	}),
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
