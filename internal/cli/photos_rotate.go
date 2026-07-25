package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var photosRotateCmd = &cobra.Command{
	Use:   "rotate [photo-id]",
	Short: "Rotate photo",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.rotate", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		degrees, _ := ctx.Cmd.Flags().GetInt("degrees")
		return ctx.runMutation(mutationSpec{
			Command:  "photos.rotate",
			Method:   "flickr.photos.transform.rotate",
			Resource: map[string]any{"photo_id": photoID, "degrees": degrees},
			PlanMsg:  fmt.Sprintf("Would rotate photo %s by %d degrees\n", photoID, degrees),
			PlanData: map[string]any{"photo_id": photoID, "degrees": degrees},
		}, func() (any, error) {
			if degrees != 90 && degrees != 180 && degrees != 270 {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrValidationFailed, "degrees must be 90, 180, or 270"))
			}
			params := map[string]string{"photo_id": photoID, "degrees": fmt.Sprintf("%d", degrees)}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.transform.rotate", params, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Rotated photo %s by %d degrees\n", photoID, degrees)
			return map[string]any{"photo_id": photoID, "degrees": degrees}, nil
		})
	}),
}
