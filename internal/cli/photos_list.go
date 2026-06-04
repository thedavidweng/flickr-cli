package cli

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var photosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your photos",
	RunE: withAuth("photos.list", func(ctx *CmdContext) error {
		page, _ := ctx.Cmd.Flags().GetInt("page")
		perPage, _ := ctx.Cmd.Flags().GetInt("per-page")

		params := map[string]string{
			"user_id":  "me",
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var result flickr.PhotoListResponse

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.people.getPhotos", params, &result); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		photos := make([]model.PhotoSummary, len(result.Photos.Photo))
		for i, p := range result.Photos.Photo {
			photos[i] = model.PhotoSummary{
				ID:    p.ID,
				Title: p.Title,
			}
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"items": photos,
				"pagination": map[string]any{
					"page":     result.Photos.Page,
					"pages":    result.Photos.Pages,
					"per_page": result.Photos.PerPage,
					"total":    result.Photos.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(ctx.R.Out)
		tw.Header("ID", "Title")
		for _, p := range photos {
			tw.Row(p.ID, p.Title)
		}
		return tw.Flush()
	}),
}

var photosSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search photos",
	RunE: withAuth("photos.search", func(ctx *CmdContext) error {
		params := map[string]string{}
		if userID, _ := ctx.Cmd.Flags().GetString("user-id"); userID != "" {
			params["user_id"] = userID
		}
		if text, _ := ctx.Cmd.Flags().GetString("text"); text != "" {
			params["text"] = text
		}
		if tags, _ := ctx.Cmd.Flags().GetStringSlice("tag"); len(tags) > 0 {
			tagsStr := ""
			for i, t := range tags {
				if i > 0 {
					tagsStr += ","
				}
				tagsStr += t
			}
			params["tags"] = tagsStr
		}
		if machineTags, _ := ctx.Cmd.Flags().GetStringSlice("machine-tag"); len(machineTags) > 0 {
			mtStr := ""
			for i, mt := range machineTags {
				if i > 0 {
					mtStr += ","
				}
				mtStr += mt
			}
			params["machine_tags"] = mtStr
		}
		if minUpload, _ := ctx.Cmd.Flags().GetString("min-upload-date"); minUpload != "" {
			params["min_upload_date"] = minUpload
		}
		if maxUpload, _ := ctx.Cmd.Flags().GetString("max-upload-date"); maxUpload != "" {
			params["max_upload_date"] = maxUpload
		}
		if minTaken, _ := ctx.Cmd.Flags().GetString("min-taken-date"); minTaken != "" {
			params["min_taken_date"] = minTaken
		}
		if maxTaken, _ := ctx.Cmd.Flags().GetString("max-taken-date"); maxTaken != "" {
			params["max_taken_date"] = maxTaken
		}
		if privacy, _ := ctx.Cmd.Flags().GetString("privacy"); privacy != "" {
			params["privacy_filter"] = privacy
		}

		page, _ := ctx.Cmd.Flags().GetInt("page")
		perPage, _ := ctx.Cmd.Flags().GetInt("per-page")
		params["page"] = fmt.Sprintf("%d", page)
		params["per_page"] = fmt.Sprintf("%d", perPage)

		if extras, _ := ctx.Cmd.Flags().GetString("extras"); extras != "" {
			params["extras"] = extras
		}

		var result flickr.PhotoSearchResponse

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.search", params, &result); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		photos := make([]model.PhotoSummary, len(result.Photos.Photo))
		for i, p := range result.Photos.Photo {
			photos[i] = model.PhotoSummary{
				ID:    p.ID,
				Title: p.Title,
				Owner: p.Owner,
			}
		}

		if ctx.App.JSON {
			return ctx.R.Success(ctx.Meta, map[string]any{
				"items": photos,
				"pagination": map[string]any{
					"page":     result.Photos.Page,
					"pages":    result.Photos.Pages,
					"per_page": result.Photos.PerPage,
					"total":    result.Photos.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(ctx.R.Out)
		tw.Header("ID", "Title", "Owner")
		for _, p := range photos {
			tw.Row(p.ID, p.Title, p.Owner)
		}
		return tw.Flush()
	}),
}

var photosShowCmd = &cobra.Command{
	Use:   "show [photo-id]",
	Short: "Show photo metadata",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.show", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		var warnings []string

		infoParams := map[string]string{"photo_id": photoID}
		var infoResult struct {
			Photo struct {
				ID    string `json:"id"`
				Title struct {
					Content string `json:"_content"`
				} `json:"title"`
				Description struct {
					Content string `json:"_content"`
				} `json:"description"`
				Owner struct {
					NSID string `json:"nsid"`
				} `json:"owner"`
				Tags struct {
					Tag []struct {
						ID      string `json:"id"`
						Raw     string `json:"raw"`
						Machine int    `json:"machine"`
					} `json:"tag"`
				} `json:"tags"`
			} `json:"photo"`
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.getInfo", infoParams, &infoResult); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "getInfo: %v", err))
		}

		sizeParams := map[string]string{"photo_id": photoID}
		var sizeResult struct {
			Sizes struct {
				Size []struct {
					Label  string `json:"label"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
					Source string `json:"source"`
					URL    string `json:"url"`
				} `json:"size"`
			} `json:"sizes"`
		}
		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.getSizes", sizeParams, &sizeResult); err != nil {
			warnings = append(warnings, fmt.Sprintf("getSizes failed: %v", err))
		}

		ctxParams := map[string]string{"photo_id": photoID}
		var ctxResult struct {
			Set []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"set"`
		}
		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.getAllContexts", ctxParams, &ctxResult); err != nil {
			warnings = append(warnings, fmt.Sprintf("getAllContexts failed: %v", err))
		}

		photo := map[string]any{
			"id":          infoResult.Photo.ID,
			"title":       infoResult.Photo.Title.Content,
			"description": infoResult.Photo.Description.Content,
			"owner":       infoResult.Photo.Owner.NSID,
		}

		tags := make([]string, len(infoResult.Photo.Tags.Tag))
		for i, t := range infoResult.Photo.Tags.Tag {
			tags[i] = t.Raw
		}
		photo["tags"] = tags

		if len(sizeResult.Sizes.Size) > 0 {
			sizes := make([]map[string]any, len(sizeResult.Sizes.Size))
			for i, s := range sizeResult.Sizes.Size {
				sizes[i] = map[string]any{
					"label":  s.Label,
					"width":  s.Width,
					"height": s.Height,
					"source": s.Source,
					"url":    s.URL,
				}
			}
			photo["sizes"] = sizes
		}

		if len(ctxResult.Set) > 0 {
			albums := make([]map[string]string, len(ctxResult.Set))
			for i, s := range ctxResult.Set {
				albums[i] = map[string]string{"id": s.ID, "title": s.Title}
			}
			photo["albums"] = albums
		}

		return ctx.R.Success(ctx.Meta, photo, warnings)
	}),
}
