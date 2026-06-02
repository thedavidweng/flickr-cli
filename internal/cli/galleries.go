package cli

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var galleriesCmd = &cobra.Command{
	Use:   "galleries",
	Short: "Manage galleries",
}

var galleriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List galleries",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "galleries.list",
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
			Galleries struct {
				Gallery []struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Photos      int    `json:"count_photos"`
				} `json:"gallery"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"galleries"`
		}

		if err := client.Call(cmd.Context(), "flickr.galleries.getList", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"galleries": result.Galleries.Gallery,
				"page":      result.Galleries.Page,
				"pages":     result.Galleries.Pages,
				"total":     result.Galleries.Total,
			}, nil)
		}

		for _, g := range result.Galleries.Gallery {
			r.Human("%s\t%s (%d photos)\n", g.ID, g.Title, g.Photos)
		}
		return nil
	},
}

var galleriesPhotosCmd = &cobra.Command{
	Use:   "photos [gallery-id]",
	Short: "List photos in a gallery",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "galleries.photos",
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
			"gallery_id": args[0],
			"page":       fmt.Sprintf("%d", page),
			"per_page":   fmt.Sprintf("%d", perPage),
		}

		var result struct {
			Gallery struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"gallery"`
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

		if err := client.Call(cmd.Context(), "flickr.galleries.getPhotos", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"gallery": result.Gallery,
				"photos":  result.Photos.Photo,
				"page":    result.Photos.Page,
				"pages":   result.Photos.Pages,
				"total":   result.Photos.Total,
			}, nil)
		}

		for _, p := range result.Photos.Photo {
			r.Human("%s\t%s\n", p.ID, p.Title)
		}
		return nil
	},
}

func init() {
	galleriesListCmd.Flags().Int("page", 1, "page number")
	galleriesListCmd.Flags().Int("per-page", 50, "items per page")
	galleriesPhotosCmd.Flags().Int("page", 1, "page number")
	galleriesPhotosCmd.Flags().Int("per-page", 50, "items per page")

	galleriesCmd.AddCommand(galleriesListCmd)
	galleriesCmd.AddCommand(galleriesPhotosCmd)
}
