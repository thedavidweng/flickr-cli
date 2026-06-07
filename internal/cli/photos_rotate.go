package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

var photosRotateCmd = &cobra.Command{
	Use:   "rotate [photo-id]",
	Short: "Rotate photo",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.rotate", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]

		gateResult := safety.Check(safety.GateInput{ReadOnly: ctx.App.ReadOnly, DryRun: ctx.App.DryRun, Confirm: ctx.App.Confirm}, safety.Mutation{
			Command: "photos.rotate",
			Method:  "flickr.photos.transform.rotate",
			Risk:    safety.ClassifyRisk("photos.rotate"),
		})
		if gateResult.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gateResult.Error)
		}
		if gateResult.Planned {
			degrees, _ := ctx.Cmd.Flags().GetInt("degrees")
			ctx.R.Human("Would rotate photo %s by %d degrees\n", photoID, degrees)
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_id": photoID, "degrees": degrees}, nil)
		}

		degrees, _ := ctx.Cmd.Flags().GetInt("degrees")
		if degrees != 90 && degrees != 180 && degrees != 270 {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "degrees must be 90, 180, or 270"))
		}

		params := map[string]string{
			"photo_id": photoID,
			"degrees":  fmt.Sprintf("%d", degrees),
		}

		if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.transform.rotate", params, nil); err != nil {
			return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		ctx.R.Human("Rotated photo %s by %d degrees\n", photoID, degrees)
		return ctx.R.Success(ctx.Meta, map[string]any{"photo_id": photoID, "degrees": degrees}, nil)
	}),
}
