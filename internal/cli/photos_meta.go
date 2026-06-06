package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

var photosSetMetaCmd = &cobra.Command{
	Use:   "set-meta [photo-id]",
	Short: "Set photo title and description",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-meta", func(ctx *CmdContext) error {
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "photos.set-meta",
			Method:  "flickr.photos.setMeta",
			Risk:    safety.ClassifyRisk("photos.set-meta"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			photoID := ctx.Args[0]
			title, _ := ctx.Cmd.Flags().GetString("title")
			description, _ := ctx.Cmd.Flags().GetString("description")
			ctx.R.Human("Would set metadata on photo %s (title=%q, description=%q)\n", photoID, title, description)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": photoID, "title": title, "description": description}, nil)
		}

		photoID := ctx.Args[0]
		title, _ := ctx.Cmd.Flags().GetString("title")
		description, _ := ctx.Cmd.Flags().GetString("description")

		params := map[string]string{
			"photo_id": photoID,
			"title":    title,
		}
		if description != "" {
			params["description"] = description
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.setMeta", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Updated metadata for photo %s\n", photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID}, nil)
	}),
}

var photosSetTagsCmd = &cobra.Command{
	Use:   "set-tags [photo-id]",
	Short: "Set photo tags (replaces existing)",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-tags", func(ctx *CmdContext) error {
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "photos.set-tags",
			Method:  "flickr.photos.setTags",
			Risk:    safety.ClassifyRisk("photos.set-tags"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			ctx.R.Human("Would set tags on photo %s\n", ctx.Args[0])
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": ctx.Args[0]}, nil)
		}

		photoID := ctx.Args[0]
		tagSlice, _ := ctx.Cmd.Flags().GetStringSlice("tag")
		csvTags, _ := ctx.Cmd.Flags().GetString("tags")
		if csvTags != "" {
			for _, t := range strings.Split(csvTags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tagSlice = append(tagSlice, t)
				}
			}
		}
		tags := strings.Join(tagSlice, " ")

		params := map[string]string{
			"photo_id": photoID,
			"tags":     tags,
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.setTags", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Set tags on photo %s\n", photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "tags": tagSlice}, nil)
	}),
}

var photosAddTagsCmd = &cobra.Command{
	Use:   "add-tags [photo-id]",
	Short: "Add tags to photo",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.add-tags", func(ctx *CmdContext) error {
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "photos.add-tags",
			Method:  "flickr.photos.addTags",
			Risk:    safety.ClassifyRisk("photos.add-tags"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			ctx.R.Human("Would add tags to photo %s\n", ctx.Args[0])
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": ctx.Args[0]}, nil)
		}

		photoID := ctx.Args[0]
		tagSlice, _ := ctx.Cmd.Flags().GetStringSlice("tag")
		csvTags, _ := ctx.Cmd.Flags().GetString("tags")
		if csvTags != "" {
			for _, t := range strings.Split(csvTags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tagSlice = append(tagSlice, t)
				}
			}
		}
		tags := strings.Join(tagSlice, " ")

		params := map[string]string{
			"photo_id": photoID,
			"tags":     tags,
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.addTags", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Added tags to photo %s\n", photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "tags": tagSlice}, nil)
	}),
}

var photosRemoveTagCmd = &cobra.Command{
	Use:   "remove-tag [photo-id]",
	Short: "Remove a tag from photo",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.remove-tag", func(ctx *CmdContext) error {
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "photos.remove-tag",
			Method:  "flickr.photos.removeTag",
			Risk:    safety.ClassifyRisk("photos.remove-tag"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			tagID, _ := ctx.Cmd.Flags().GetString("tag-id")
			ctx.R.Human("Would remove tag %s from photo %s\n", tagID, ctx.Args[0])
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": ctx.Args[0], "tag_id": tagID}, nil)
		}

		photoID := ctx.Args[0]
		tagID, _ := ctx.Cmd.Flags().GetString("tag-id")

		params := map[string]string{
			"photo_id": photoID,
			"tag_id":   tagID,
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.removeTag", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Removed tag %s from photo %s\n", tagID, photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "tag_id": tagID}, nil)
	}),
}
