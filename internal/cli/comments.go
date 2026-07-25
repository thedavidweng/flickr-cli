package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage photo comments",
}

var commentsListCmd = &cobra.Command{
	Use:   "list [photo-id]",
	Short: "List comments on a photo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "comments.list",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		params := map[string]string{
			"photo_id": args[0],
		}

		var result struct {
			Comments struct {
				Comment []struct {
					ID      string `json:"id"`
					Author  string `json:"authorname"`
					Content string `json:"_content"`
					Date    string `json:"datecreate"`
				} `json:"comment"`
			} `json:"comments"`
		}

		if err := client.Call(cmd.Context(), "flickr.photos.comments.getList", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"comments": result.Comments.Comment,
			}, nil)
		}

		for _, c := range result.Comments.Comment {
			r.Human("[%s] %s: %s\n", c.ID, c.Author, c.Content)
		}
		return nil
	},
}

var commentsAddCmd = &cobra.Command{
	Use:   "add [photo-id] [text]",
	Short: "Add a comment to a photo",
	Args:  cobra.ExactArgs(2),
	RunE: withAuth("comments.add", func(ctx *CmdContext) error {
		photoID, text := ctx.Args[0], ctx.Args[1]
		return ctx.runMutation(mutationSpec{
			Command:  "comments.add",
			Method:   "flickr.photos.comments.addComment",
			Resource: map[string]any{"photo_id": photoID},
			PlanMsg:  fmt.Sprintf("Would add comment to photo %s\n", photoID),
		}, func() (any, error) {
			params := map[string]string{"photo_id": photoID, "comment_text": text}
			var result struct {
				Comment struct {
					ID string `json:"id"`
				} `json:"comment"`
			}
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.comments.addComment", params, &result); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Added comment %s to photo %s\n", result.Comment.ID, photoID)
			return map[string]any{"comment_id": result.Comment.ID}, nil
		})
	}),
}

var commentsDeleteCmd = &cobra.Command{
	Use:   "delete [comment-id]",
	Short: "Delete a comment",
	Args:  cobra.ExactArgs(1),
	RunE: withAuth("comments.delete", func(ctx *CmdContext) error {
		commentID := ctx.Args[0]
		return ctx.runMutation(mutationSpec{
			Command:  "comments.delete",
			Method:   "flickr.photos.comments.deleteComment",
			Resource: map[string]any{"comment_id": commentID},
			PlanMsg:  fmt.Sprintf("Would delete comment %s\n", commentID),
		}, func() (any, error) {
			if err := ctx.Client.Call(ctx.Cmd.Context(), "flickr.photos.comments.deleteComment", map[string]string{"comment_id": commentID}, nil); err != nil {
				return nil, ctx.R.Failure(ctx.Meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
			}
			ctx.R.Human("Deleted comment %s\n", commentID)
			return map[string]any{"comment_id": commentID}, nil
		})
	}),
}

func init() {
	commentsCmd.AddCommand(commentsListCmd)
	commentsCmd.AddCommand(commentsAddCmd)
	commentsCmd.AddCommand(commentsDeleteCmd)
}
