package cli

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "List files in albums",
}

var filesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List photo IDs in albums",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "files.list",
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

		albumIDs, _ := cmd.Flags().GetStringSlice("album-id")
		albumTitles, _ := cmd.Flags().GetStringSlice("album")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		// If --album titles are provided, resolve them to IDs via flickr.photosets.getList
		if len(albumTitles) > 0 && len(albumIDs) == 0 {
			listParams := map[string]string{
				"user_id":  "me",
				"page":     "1",
				"per_page": "500",
			}

			var listResult struct {
				Photosets struct {
					Photoset []struct {
						ID    string `json:"id"`
						Title struct {
							Content string `json:"_content"`
						} `json:"title"`
					} `json:"photoset"`
				} `json:"photosets"`
			}

			if err := client.Call(cmd.Context(), "flickr.photosets.getList", listParams, &listResult); err != nil {
				return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}

			titleToID := make(map[string]string)
			for _, ps := range listResult.Photosets.Photoset {
				titleToID[ps.Title.Content] = ps.ID
			}

			for _, title := range albumTitles {
				id, ok := titleToID[title]
				if !ok {
					return r.Failure(meta, output.Errorf(model.ErrResourceNotFound, "album %q not found", title))
				}
				albumIDs = append(albumIDs, id)
			}
		}

		if len(albumIDs) == 0 {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "at least one --album or --album-id is required"))
		}

		// Fetch photos from each album via flickr.photosets.getPhotos
		allPhotoIDs := make([]string, 0)
		for _, albumID := range albumIDs {
			params := map[string]string{
				"photoset_id": albumID,
				"page":        fmt.Sprintf("%d", page),
				"per_page":    fmt.Sprintf("%d", perPage),
			}

			var result struct {
				Photoset struct {
					Photo []struct {
						ID    string `json:"id"`
						Title string `json:"title"`
					} `json:"photo"`
					Page    int `json:"page"`
					Pages   int `json:"pages"`
					PerPage int `json:"perpage"`
					Total   int `json:"total"`
				} `json:"photoset"`
			}

			if err := client.Call(cmd.Context(), "flickr.photosets.getPhotos", params, &result); err != nil {
				return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}

			for _, p := range result.Photoset.Photo {
				allPhotoIDs = append(allPhotoIDs, p.ID)
			}
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"items": allPhotoIDs,
			}, nil)
		}

		for _, id := range allPhotoIDs {
			r.Human("%s\n", id)
		}
		return nil
	},
}

func init() {
	filesListCmd.Flags().StringSlice("album", nil, "album title (repeatable)")
	filesListCmd.Flags().StringSlice("album-id", nil, "album ID (repeatable)")
	filesListCmd.Flags().Int("page", 1, "page number")
	filesListCmd.Flags().Int("per-page", 50, "items per page")

	filesCmd.AddCommand(filesListCmd)
}
