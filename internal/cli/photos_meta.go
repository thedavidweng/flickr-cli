package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var photosSetMetaCmd = &cobra.Command{
	Use:   "set-meta [photo-id]",
	Short: "Set photo title and description",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-meta", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		title, _ := ctx.Cmd.Flags().GetString("title")
		description, _ := ctx.Cmd.Flags().GetString("description")
		return ctx.runMutation(mutationSpec{
			Command:  "photos.set-meta",
			Method:   "flickr.photos.setMeta",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would set metadata on photo %s (title=%q, description=%q)\n", photoID, title, description),
			PlanData: map[string]any{"photo_id": photoID, "title": title, "description": description},
		}, func() (any, error) {
			params := map[string]string{"photo_id": photoID, "title": title}
			if description != "" {
				params["description"] = description
			}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.setMeta", params, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Updated metadata for photo %s\n", photoID)
			return map[string]any{"photo_id": photoID}, nil
		})
	}),
}

var photosSetTagsCmd = &cobra.Command{
	Use:   "set-tags [photo-id]",
	Short: "Set photo tags (replaces existing)",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-tags", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		tagSlice := collectTags(ctx)
		return ctx.runMutation(mutationSpec{
			Command:  "photos.set-tags",
			Method:   "flickr.photos.setTags",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would set tags on photo %s\n", photoID),
			PlanData: map[string]any{"photo_id": photoID},
		}, func() (any, error) {
			params := map[string]string{"photo_id": photoID, "tags": strings.Join(tagSlice, " ")}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.setTags", params, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Set tags on photo %s\n", photoID)
			return map[string]any{"photo_id": photoID, "tags": tagSlice}, nil
		})
	}),
}

var photosAddTagsCmd = &cobra.Command{
	Use:   "add-tags [photo-id]",
	Short: "Add tags to photo",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.add-tags", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		tagSlice := collectTags(ctx)
		return ctx.runMutation(mutationSpec{
			Command:  "photos.add-tags",
			Method:   "flickr.photos.addTags",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would add tags to photo %s\n", photoID),
			PlanData: map[string]any{"photo_id": photoID},
		}, func() (any, error) {
			params := map[string]string{"photo_id": photoID, "tags": strings.Join(tagSlice, " ")}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.addTags", params, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Added tags to photo %s\n", photoID)
			return map[string]any{"photo_id": photoID, "tags": tagSlice}, nil
		})
	}),
}

var photosRemoveTagCmd = &cobra.Command{
	Use:   "remove-tag [photo-id]",
	Short: "Remove a tag from photo",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.remove-tag", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		tagID, _ := ctx.Cmd.Flags().GetString("tag-id")
		return ctx.runMutation(mutationSpec{
			Command:  "photos.remove-tag",
			Method:   "flickr.photos.removeTag",
			Resource: map[string]any{"photo_id": photoID, "tag_id": tagID},
			PlanMsg:  fmt.Sprintf("Would remove tag %s from photo %s\n", tagID, photoID),
			PlanData: map[string]any{"photo_id": photoID, "tag_id": tagID},
		}, func() (any, error) {
			params := map[string]string{"photo_id": photoID, "tag_id": tagID}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.removeTag", params, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Removed tag %s from photo %s\n", tagID, photoID)
			return map[string]any{"photo_id": photoID, "tag_id": tagID}, nil
		})
	}),
}

func collectTags(ctx *CmdContext) []string {
	tagSlice, _ := ctx.Cmd.Flags().GetStringSlice("tag")
	csvTags, _ := ctx.Cmd.Flags().GetString("tags")
	if csvTags != "" {
		for _, t := range strings.Split(csvTags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tagSlice = append(tagSlice, t)
			}
		}
	}
	return tagSlice
}
