package cli

import (
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
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
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
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
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "comments.add",
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
			"comment_text": args[1],
		}

		var result struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		}

		if err := client.Call(cmd.Context(), "flickr.photos.comments.addComment", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Added comment %s to photo %s\n", result.Comment.ID, args[0])
		return r.Success(meta, map[string]any{"comment_id": result.Comment.ID}, nil)
	},
}

var commentsDeleteCmd = &cobra.Command{
	Use:   "delete [comment-id]",
	Short: "Delete a comment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "comments.delete",
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

		if !app.Confirm {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrConfirmationRequired,
				"Use --confirm to delete this comment",
				map[string]any{"comment_id": args[0]},
			))
		}

		params := map[string]string{
			"comment_id": args[0],
		}

		if err := client.Call(cmd.Context(), "flickr.photos.comments.deleteComment", params, nil); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		r.Human("Deleted comment %s\n", args[0])
		return r.Success(meta, map[string]any{"comment_id": args[0]}, nil)
	},
}

func init() {
	commentsCmd.AddCommand(commentsListCmd)
	commentsCmd.AddCommand(commentsAddCmd)
	commentsCmd.AddCommand(commentsDeleteCmd)
}
