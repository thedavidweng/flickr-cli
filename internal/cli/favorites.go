package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage favorite photos",
}

var favoritesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List favorite photos",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "favorites.list",
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

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		params := map[string]string{
			"user_id":  "me",
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var result struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"photo"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.favorites.getList", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		photos := make([]model.PhotoSummary, len(result.Photos.Photo))
		for i, p := range result.Photos.Photo {
			photos[i] = model.PhotoSummary{ID: p.ID, Title: p.Title}
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"photos": photos,
				"page":   result.Photos.Page,
				"pages":  result.Photos.Pages,
				"total":  result.Photos.Total,
			}, nil)
		}

		for _, p := range photos {
			r.Human("%s\t%s\n", p.ID, p.Title)
		}
		return nil
	},
}

var favoritesAddCmd = &cobra.Command{
	Use:   "add [photo-id]",
	Short: "Add a photo to favorites",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("favorites.add", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		return ctx.runMutation(mutationSpec{
			Command:  "favorites.add",
			Method:   "flickr.favorites.add",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would add %s to favorites\n", photoID),
		}, func() (any, error) {
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.favorites.add", map[string]string{"photo_id": photoID}, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Added %s to favorites\n", photoID)
			return map[string]any{"photo_id": photoID}, nil
		})
	}),
}

var favoritesRemoveCmd = &cobra.Command{
	Use:   "remove [photo-id]",
	Short: "Remove a photo from favorites",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("favorites.remove", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		return ctx.runMutation(mutationSpec{
			Command:  "favorites.remove",
			Method:   "flickr.favorites.remove",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would remove %s from favorites\n", photoID),
		}, func() (any, error) {
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.favorites.remove", map[string]string{"photo_id": photoID}, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Removed %s from favorites\n", photoID)
			return map[string]any{"photo_id": photoID}, nil
		})
	}),
}

func init() {
	favoritesListCmd.Flags().Int("page", 1, "page number")
	favoritesListCmd.Flags().Int("per-page", 50, "items per page")

	favoritesCmd.AddCommand(favoritesListCmd)
	favoritesCmd.AddCommand(favoritesAddCmd)
	favoritesCmd.AddCommand(favoritesRemoveCmd)
}
