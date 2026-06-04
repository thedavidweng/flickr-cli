package cli

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
	"github.com/spf13/cobra"
)

var photosSetLocationCmd = &cobra.Command{
	Use:   "set-location [photo-id]",
	Short: "Set photo location",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-location", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]

		gateResult := safety.Check(safety.GateInput{ReadOnly: ctx.App.ReadOnly, DryRun: ctx.App.DryRun, Confirm: ctx.App.Confirm}, safety.Mutation{
			Command: "photos.set-location",
			Method:  "flickr.photos.geo.setLocation",
			Risk:    safety.ClassifyRisk("photos.set-location"),
		})
		if gateResult.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gateResult.Error)
		}
		if gateResult.Planned {
			ctx.R.Human("Would set location for photo %s\n", photoID)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": photoID}, nil)
		}

		lat, _ := ctx.Cmd.Flags().GetFloat64("lat")
		lon, _ := ctx.Cmd.Flags().GetFloat64("lon")
		accuracy, _ := ctx.Cmd.Flags().GetInt("accuracy")
		context, _ := ctx.Cmd.Flags().GetInt("context")

		params := map[string]string{
			"photo_id": photoID,
			"lat":      fmt.Sprintf("%f", lat),
			"lon":      fmt.Sprintf("%f", lon),
		}
		if accuracy > 0 {
			params["accuracy"] = fmt.Sprintf("%d", accuracy)
		}
		if context > 0 {
			params["context"] = fmt.Sprintf("%d", context)
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.geo.setLocation", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Set location for photo %s\n", photoID)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "lat": lat, "lon": lon}, nil)
	}),
}
