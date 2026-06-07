package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage contacts",
}

var contactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your contacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "contacts.list",
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

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		params := map[string]string{
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var result struct {
			Contacts struct {
				Contact []struct {
					ID       string `json:"nsid"`
					Username string `json:"username"`
					RealName string `json:"realname"`
					Friend   bool   `json:"friend"`
					Family   bool   `json:"family"`
				} `json:"contact"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"contacts"`
		}

		if err := client.Call(cmd.Context(), "flickr.contacts.getList", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"contacts": result.Contacts.Contact,
				"page":     result.Contacts.Page,
				"pages":    result.Contacts.Pages,
				"total":    result.Contacts.Total,
			}, nil)
		}

		for _, c := range result.Contacts.Contact {
			r.Human("%s\t%s\n", c.ID, c.RealName)
		}
		return nil
	},
}

func init() {
	contactsListCmd.Flags().Int("page", 1, "page number")
	contactsListCmd.Flags().Int("per-page", 50, "items per page")

	contactsCmd.AddCommand(contactsListCmd)
}
