package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

var photosDeleteCmd = &cobra.Command{
	Use:   "delete [photo-id...]",
	Short: "Delete photos",
	Args:  cobra.MinimumNArgs(1),
	RunE: withAuth("photos.delete", func(ctx *CmdContext) error {
		gate := safety.Check(safety.GateInput{
			ReadOnly: ctx.App.ReadOnly,
			DryRun:   ctx.App.DryRun,
			Confirm:  ctx.App.Confirm,
		}, safety.Mutation{
			Command: "photos.delete",
			Method:  "flickr.photos.delete",
			Risk:    safety.ClassifyRisk("photos.delete"),
		})
		if gate.Error != nil {
			return ctx.R.Failure(ctx.Meta, *gate.Error)
		}
		if gate.Planned {
			ctx.R.Human("Would delete %d photo(s): %s\n", len(ctx.Args), strings.Join(ctx.Args, ", "))
			return ctx.R.Success(ctx.Meta, map[string]any{"planned": true, "photo_ids": ctx.Args}, nil)
		}

		var deleted []string
		for _, id := range ctx.Args {
			params := map[string]string{"photo_id": id}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.delete", params, nil); err != nil {
				return ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "delete %s: %v", id, err))
			}
			deleted = append(deleted, id)
		}

		ctx.R.Human("Deleted %d photo(s): %s\n", len(deleted), strings.Join(deleted, ", "))
		return ctx.R.Success(ctx.Meta, map[string]any{"deleted": deleted}, nil)
	}),
}
