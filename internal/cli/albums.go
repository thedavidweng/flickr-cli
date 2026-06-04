package cli

import (
	"fmt"
	"sort"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
	"github.com/spf13/cobra"
)

var albumsCmd = &cobra.Command{
	Use:   "albums",
	Short: "Manage Flickr albums/photosets",
}

var albumsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List albums",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "albums.list",
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
		sortBy, _ := cmd.Flags().GetString("sort")

		params := map[string]string{
			"user_id":  "me",
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var result struct {
			Photosets struct {
				Photoset []struct {
					ID    string `json:"id"`
					Title struct {
						Content string `json:"_content"`
					} `json:"title"`
					Description struct {
						Content string `json:"_content"`
					} `json:"description"`
					Photos         int    `json:"photos"`
					PrimaryPhotoID string `json:"primary_photo_id"`
					DateCreate     string `json:"date_create"`
					DateUpdate     string `json:"date_update"`
				} `json:"photoset"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photosets"`
		}

		if err := client.Call(cmd.Context(), "flickr.photosets.getList", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		albums := make([]model.Album, len(result.Photosets.Photoset))
		for i, ps := range result.Photosets.Photoset {
			albums[i] = model.Album{
				ID:             ps.ID,
				Title:          ps.Title.Content,
				Description:    ps.Description.Content,
				PhotoCount:     ps.Photos,
				PrimaryPhotoID: ps.PrimaryPhotoID,
				CreatedAt:      ps.DateCreate,
				UpdatedAt:      ps.DateUpdate,
			}
		}

		// Sort locally
		switch sortBy {
		case "title":
			sort.Slice(albums, func(i, j int) bool { return albums[i].Title < albums[j].Title })
		case "created":
			sort.Slice(albums, func(i, j int) bool { return albums[i].CreatedAt < albums[j].CreatedAt })
		case "updated":
			sort.Slice(albums, func(i, j int) bool { return albums[i].UpdatedAt < albums[j].UpdatedAt })
		case "count":
			sort.Slice(albums, func(i, j int) bool { return albums[i].PhotoCount > albums[j].PhotoCount })
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"items": albums,
				"pagination": map[string]any{
					"page":     result.Photosets.Page,
					"pages":    result.Photosets.Pages,
					"per_page": result.Photosets.PerPage,
					"total":    result.Photosets.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(r.Out)
		tw.Header("ID", "Title", "Photos")
		for _, a := range albums {
			tw.Row(a.ID, a.Title, fmt.Sprintf("%d", a.PhotoCount))
		}
		return tw.Flush()
	},
}

var albumsShowCmd = &cobra.Command{
	Use:   "show [album-id]",
	Short: "Show album metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "albums.show",
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

		albumID := args[0]
		params := map[string]string{
			"photoset_id": albumID,
		}

		var result struct {
			Photoset struct {
				ID    string `json:"id"`
				Title struct {
					Content string `json:"_content"`
				} `json:"title"`
				Description struct {
					Content string `json:"_content"`
				} `json:"description"`
				Photos         int    `json:"photos"`
				PrimaryPhotoID string `json:"primary_photo_id"`
				DateCreate     string `json:"date_create"`
				DateUpdate     string `json:"date_update"`
			} `json:"photoset"`
		}

		if err := client.Call(cmd.Context(), "flickr.photosets.getInfo", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		album := model.Album{
			ID:             result.Photoset.ID,
			Title:          result.Photoset.Title.Content,
			Description:    result.Photoset.Description.Content,
			PhotoCount:     result.Photoset.Photos,
			PrimaryPhotoID: result.Photoset.PrimaryPhotoID,
			CreatedAt:      result.Photoset.DateCreate,
			UpdatedAt:      result.Photoset.DateUpdate,
		}

		if app.JSON {
			return r.Success(meta, album, nil)
		}

		r.Human("ID:          %s\n", album.ID)
		r.Human("Title:       %s\n", album.Title)
		r.Human("Description: %s\n", album.Description)
		r.Human("Photos:      %d\n", album.PhotoCount)
		r.Human("Created:     %s\n", album.CreatedAt)
		r.Human("Updated:     %s\n", album.UpdatedAt)
		return nil
	},
}

var albumsPhotosCmd = &cobra.Command{
	Use:   "photos [album-id]",
	Short: "List photos in an album",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "albums.photos",
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

		albumID := args[0]
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

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

		photos := make([]model.PhotoSummary, len(result.Photoset.Photo))
		for i, p := range result.Photoset.Photo {
			photos[i] = model.PhotoSummary{
				ID:    p.ID,
				Title: p.Title,
			}
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"items": photos,
				"pagination": map[string]any{
					"page":     result.Photoset.Page,
					"pages":    result.Photoset.Pages,
					"per_page": result.Photoset.PerPage,
					"total":    result.Photoset.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(r.Out)
		tw.Header("ID", "Title")
		for _, p := range photos {
			tw.Row(p.ID, p.Title)
		}
		return tw.Flush()
	},
}

var albumsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new album",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "albums.create",
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

		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		primaryPhotoID, _ := cmd.Flags().GetString("primary-photo-id")

		if title == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "title is required"))
		}
		if primaryPhotoID == "" {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "primary-photo-id is required"))
		}

		if app.DryRun {
			r.Human("Would create album %q with primary photo %s\n", title, primaryPhotoID)
			return r.Success(meta, map[string]any{"planned": true, "title": title, "primary_photo_id": primaryPhotoID}, nil)
		}

		params := map[string]string{
			"title":            title,
			"primary_photo_id": primaryPhotoID,
		}
		if description != "" {
			params["description"] = description
		}

		var result struct {
			Photoset struct {
				ID string `json:"id"`
			} `json:"photoset"`
			Stat string `json:"stat"`
		}

		if err := client.Call(cmd.Context(), "flickr.photosets.create", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Created album %s\n", result.Photoset.ID)
		return r.Success(meta, map[string]any{"id": result.Photoset.ID, "title": title}, nil)
	},
}

var albumsUpdateCmd = &cobra.Command{
	Use:   "update [album-id]",
	Short: "Update album metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "albums.update",
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

		albumID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		primaryPhotoID, _ := cmd.Flags().GetString("primary-photo-id")

		if app.DryRun {
			r.Human("Would update album %s\n", albumID)
			return r.Success(meta, map[string]any{"planned": true, "id": albumID}, nil)
		}

		params := map[string]string{
			"photoset_id": albumID,
		}
		if title != "" {
			params["title"] = title
		}
		if description != "" {
			params["description"] = description
		}
		if primaryPhotoID != "" {
			params["primary_photo_id"] = primaryPhotoID
		}

		if err := client.Call(cmd.Context(), "flickr.photosets.editMeta", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Updated album %s\n", albumID)
		return r.Success(meta, map[string]any{"id": albumID}, nil)
	},
}

var albumsDeleteCmd = &cobra.Command{
	Use:   "delete [album-id]",
	Short: "Delete an album",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "albums.delete",
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

		albumID := args[0]

		if app.ReadOnly {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrReadOnlyViolation,
				"Operation blocked by --read-only flag",
				map[string]any{"command": "albums.delete", "flag": "--read-only"},
			))
		}

		if !app.Confirm {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrConfirmationRequired,
				"Use --confirm to delete album",
				map[string]any{"album_id": albumID},
			))
		}

		if app.DryRun {
			r.Human("Would delete album %s\n", albumID)
			return r.Success(meta, map[string]any{"planned": true, "id": albumID}, nil)
		}

		params := map[string]string{
			"photoset_id": albumID,
		}

		if err := client.Call(cmd.Context(), "flickr.photosets.delete", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Deleted album %s\n", albumID)
		return r.Success(meta, map[string]any{"id": albumID}, nil)
	},
}

var albumsAddPhotosCmd = &cobra.Command{
	Use:   "add-photos [album-id]",
	Short: "Add photos to an album",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("albums.add-photos", func(ctx *CmdContext) error {
		albumID := ctx.Args[0]
		photoIDs, _ := ctx.Cmd.Flags().GetStringSlice("photo-id")
		if len(photoIDs) == 0 {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "--photo-id is required"))
		}

		// Safety gate (medium risk)
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "albums.add-photos",
			Method:  "flickr.photosets.addPhoto",
			Risk:    safety.ClassifyRisk("albums.add-photos"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			ctx.R.Human("Would add %d photos to album %s\n", len(photoIDs), albumID)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "album_id": albumID, "photo_ids": photoIDs}, nil)
		}

		var added []string
		var errs []map[string]any
		for _, photoID := range photoIDs {
			if err := ctx.Client.AddToAlbum(ctx.Cmd.Context(), albumID, photoID); err != nil {
				errs = append(errs, map[string]any{"photo_id": photoID, "error": err.Error()})
			} else {
				added = append(added, photoID)
			}
		}

		if len(errs) > 0 {
			ctx.R.Human("Added %d/%d photos to album %s\n", len(added), len(photoIDs), albumID)
			return ctx.R.Failure(ctx.Meta, output.ErrorWithDetails(
				model.ErrFlickrAPI,
				fmt.Sprintf("failed to add %d photos", len(errs)),
				map[string]any{"added": added, "errors": errs},
			))
		}

		ctx.R.Human("Added %d photos to album %s\n", len(added), albumID)
		return ctx.R.Success(ctx.Meta, map[string]any{"album_id": albumID, "added": added}, nil)
	}),
}

var albumsRemovePhotosCmd = &cobra.Command{
	Use:   "remove-photos [album-id]",
	Short: "Remove photos from an album",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("albums.remove-photos", func(ctx *CmdContext) error {
		albumID := ctx.Args[0]
		photoIDs, _ := ctx.Cmd.Flags().GetStringSlice("photo-id")
		if len(photoIDs) == 0 {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "--photo-id is required"))
		}

		// Safety gate (medium risk)
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "albums.remove-photos",
			Method:  "flickr.photosets.removePhoto",
			Risk:    safety.ClassifyRisk("albums.remove-photos"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			ctx.R.Human("Would remove %d photos from album %s\n", len(photoIDs), albumID)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "album_id": albumID, "photo_ids": photoIDs}, nil)
		}

		var removed []string
		var errs []map[string]any
		for _, photoID := range photoIDs {
			params := map[string]string{
				"photoset_id": albumID,
				"photo_id":    photoID,
			}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photosets.removePhoto", params, nil); err != nil {
				errs = append(errs, map[string]any{"photo_id": photoID, "error": err.Error()})
			} else {
				removed = append(removed, photoID)
			}
		}

		if len(errs) > 0 {
			ctx.R.Human("Removed %d/%d photos from album %s\n", len(removed), len(photoIDs), albumID)
			return ctx.R.Failure(ctx.Meta, output.ErrorWithDetails(
				model.ErrFlickrAPI,
				fmt.Sprintf("failed to remove %d photos", len(errs)),
				map[string]any{"removed": removed, "errors": errs},
			))
		}

		ctx.R.Human("Removed %d photos from album %s\n", len(removed), albumID)
		return ctx.R.Success(ctx.Meta, map[string]any{"album_id": albumID, "removed": removed}, nil)
	}),
}

func init() {
	albumsListCmd.Flags().Int("page", 1, "page number")
	albumsListCmd.Flags().Int("per-page", 50, "items per page")
	albumsListCmd.Flags().String("sort", "title", "sort by: title|created|updated|count")

	albumsCreateCmd.Flags().String("title", "", "album title (required)")
	albumsCreateCmd.Flags().String("description", "", "album description")
	albumsCreateCmd.Flags().String("primary-photo-id", "", "primary photo ID (required by Flickr)")

	albumsUpdateCmd.Flags().String("title", "", "album title")
	albumsUpdateCmd.Flags().String("description", "", "album description")
	albumsUpdateCmd.Flags().String("primary-photo-id", "", "primary photo ID")

	albumsPhotosCmd.Flags().Int("page", 1, "page number")
	albumsPhotosCmd.Flags().Int("per-page", 50, "items per page")
	albumsPhotosCmd.Flags().String("privacy", "", "privacy level filter")
	albumsPhotosCmd.Flags().String("extras", "", "extra fields CSV")

	albumsAddPhotosCmd.Flags().StringSlice("photo-id", nil, "photo ID to add (repeatable, required)")
	albumsRemovePhotosCmd.Flags().StringSlice("photo-id", nil, "photo ID to remove (repeatable, required)")

	albumsCmd.AddCommand(albumsListCmd)
	albumsCmd.AddCommand(albumsShowCmd)
	albumsCmd.AddCommand(albumsCreateCmd)
	albumsCmd.AddCommand(albumsUpdateCmd)
	albumsCmd.AddCommand(albumsDeleteCmd)
	albumsCmd.AddCommand(albumsPhotosCmd)
	albumsCmd.AddCommand(albumsAddPhotosCmd)
	albumsCmd.AddCommand(albumsRemovePhotosCmd)
}
