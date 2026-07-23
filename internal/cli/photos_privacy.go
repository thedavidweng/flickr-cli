package cli

import (
	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
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

		level, err := flickr.ParsePrivacyLevel(privacy)
		if err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		params := map[string]string{"photo_id": photoID}
		for k, v := range level.PermsParams() {
			params[k] = v
		}
		switch hidden {
		case "hidden":
			params["hidden"] = "2"
		case "public":
			params["hidden"] = "1"
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.setPerms", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Set privacy for photo %s\n", photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "privacy": privacy}, nil)
	}),
}
