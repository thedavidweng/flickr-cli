package cli

import (
	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var urlsCmd = &cobra.Command{
	Use:   "urls",
	Short: "URL lookup utilities",
}

var urlsLookupUserCmd = &cobra.Command{
	Use:   "lookup-user [url]",
	Short: "Look up a user by profile URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "urls.lookupUser",
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

		params := map[string]string{
			"url": args[0],
		}

		var result struct {
			User struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"user"`
		}

		if err := client.Call(cmd.Context(), "flickr.urls.lookupUser", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, result.User, nil)
		}

		r.Human("%s\t%s\n", result.User.ID, result.User.Username)
		return nil
	},
}

func init() {
	urlsCmd.AddCommand(urlsLookupUserCmd)
}
