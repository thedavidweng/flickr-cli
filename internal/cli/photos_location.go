package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var photosSetLocationCmd = &cobra.Command{
	Use:   "set-location [photo-id]",
	Short: "Set photo location",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("photos.set-location", func(ctx *CmdContext) error {
		photoID := ctx.Args[0]
		return ctx.runMutation(mutationSpec{
			Command:  "photos.set-location",
			Method:   "flickr.photos.geo.setLocation",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would set location for photo %s\n", photoID),
			PlanData: map[string]any{"photo_id": photoID},
		}, func() (any, error) {
			lat, _ := ctx.Cmd.Flags().GetFloat64("lat")
			lon, _ := ctx.Cmd.Flags().GetFloat64("lon")
			accuracy, _ := ctx.Cmd.Flags().GetInt("accuracy")
			geoContext, _ := ctx.Cmd.Flags().GetInt("context")

			params := map[string]string{
				"photo_id": photoID,
				"lat":      fmt.Sprintf("%f", lat),
				"lon":      fmt.Sprintf("%f", lon),
			}
			if accuracy > 0 {
				params["accuracy"] = fmt.Sprintf("%d", accuracy)
			}
			if geoContext > 0 {
				params["context"] = fmt.Sprintf("%d", geoContext)
			}

			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.geo.setLocation", params, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Set location for photo %s\n", photoID)
			return map[string]any{"photo_id": photoID, "lat": lat, "lon": lon}, nil
		})
	}),
}
