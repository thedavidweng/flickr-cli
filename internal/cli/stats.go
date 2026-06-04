package cli

import (
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View statistics",
}

var statsPopularCmd = &cobra.Command{
	Use:   "popular",
	Short: "Show popular photos",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "stats.popular",
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
			"user_id": "me",
		}

		var result struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Views int    `json:"views"`
				} `json:"photo"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.stats.getPopularPhotos", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"photos": result.Photos.Photo,
			}, nil)
		}

		for _, p := range result.Photos.Photo {
			r.Human("%s\t%s (%d views)\n", p.ID, p.Title, p.Views)
		}
		return nil
	},
}

func init() {
	statsCmd.AddCommand(statsPopularCmd)
}
