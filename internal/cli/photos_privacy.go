package cli

import (
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
	"github.com/spf13/cobra"
)

var photosSetPrivacyCmd = &cobra.Command{
	Use:   "set-privacy [photo-id]",
	Short: "Set photo privacy",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-privacy", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]

		gateResult := safety.Check(safety.GateInput{ReadOnly: ctx.App.ReadOnly, DryRun: ctx.App.DryRun, Confirm: ctx.App.Confirm}, safety.Mutation{
			Command: "photos.set-privacy",
			Method:  "flickr.photos.setPerms",
			Risk:    safety.ClassifyRisk("photos.set-privacy"),
		})
		if gateResult.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gateResult.Error)
		}
		if gateResult.Planned {
			ctx.R.Human("Would set privacy for photo %s\n", photoID)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": photoID}, nil)
		}

		privacy, _ := ctx.Cmd.Flags().GetString("privacy")
		hidden, _ := ctx.Cmd.Flags().GetString("hidden")

		params := map[string]string{"photo_id": photoID}
		switch privacy {
		case "public":
			params["is_public"] = "1"
			params["is_friend"] = "0"
			params["is_family"] = "0"
		case "private":
			params["is_public"] = "0"
			params["is_friend"] = "0"
			params["is_family"] = "0"
		case "friends":
			params["is_public"] = "0"
			params["is_friend"] = "1"
			params["is_family"] = "0"
		case "family":
			params["is_public"] = "0"
			params["is_friend"] = "0"
			params["is_family"] = "1"
		case "friends-family":
			params["is_public"] = "0"
			params["is_friend"] = "1"
			params["is_family"] = "1"
		}
		if hidden == "hidden" {
			params["hidden"] = "2"
		} else if hidden == "public" {
			params["hidden"] = "1"
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.setPerms", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Set privacy for photo %s\n", photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "privacy": privacy}, nil)
	}),
}
