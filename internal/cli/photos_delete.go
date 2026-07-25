package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var photosDeleteCmd = &cobra.Command{
	Use:   "delete [photo-id...]",
	Short: "Delete photos",
	Args:  cobra.MinimumNArgs(1),
	RunE: withAuth("photos.delete", func(ctx *CmdContext) error {
		return ctx.runMutation(mutationSpec{
			Command:  "photos.delete",
			Method:   "flickr.photos.delete",
			Resource: map[string]any{"photo_ids": ctx.Args},
			PlanMsg:  fmt.Sprintf("Would delete %d photo(s): %s\n", len(ctx.Args), strings.Join(ctx.Args, ", ")),
			PlanData: map[string]any{"photo_ids": ctx.Args},
		}, func() (any, error) {
			var deleted []string
			for _, id := range ctx.Args {
				params := map[string]string{"photo_id": id}
				if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.delete", params, nil); err != nil {
					return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "delete %s: %v", id, err))
				}
				deleted = append(deleted, id)
			}
			ctx.R.Human("Deleted %d photo(s): %s\n", len(deleted), strings.Join(deleted, ", "))
			return map[string]any{"deleted": deleted}, nil
		})
	}),
}
