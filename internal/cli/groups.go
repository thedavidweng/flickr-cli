package cli

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage groups",
}

var groupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List groups you belong to",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
		meta := output.RuntimeMetaInput{
			Command:   "groups.list",
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
			Groups struct {
				Group []struct {
					ID          string `json:"nsid"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Members     int    `json:"members"`
					Privacy     int    `json:"privacy"`
				} `json:"group"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"groups"`
		}

		if err := client.Call(cmd.Context(), "flickr.groups.getList", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"groups": result.Groups.Group,
				"page":   result.Groups.Page,
				"pages":  result.Groups.Pages,
				"total":  result.Groups.Total,
			}, nil)
		}

		for _, g := range result.Groups.Group {
			r.Human("%s\t%s (%d members)\n", g.ID, g.Name, g.Members)
		}
		return nil
	},
}

var groupsSearchCmd = &cobra.Command{
	Use:   "search [text]",
	Short: "Search for groups",
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
			Command:   "groups.search",
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
			"text":     args[0],
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var result struct {
			Groups struct {
				Group []struct {
					ID      string `json:"nsid"`
					Name    string `json:"name"`
					Members int    `json:"members"`
				} `json:"group"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"groups"`
		}

		if err := client.Call(cmd.Context(), "flickr.groups.search", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"groups": result.Groups.Group,
				"page":   result.Groups.Page,
				"pages":  result.Groups.Pages,
				"total":  result.Groups.Total,
			}, nil)
		}

		for _, g := range result.Groups.Group {
			r.Human("%s\t%s (%d members)\n", g.ID, g.Name, g.Members)
		}
		return nil
	},
}

func init() {
	groupsListCmd.Flags().Int("page", 1, "page number")
	groupsListCmd.Flags().Int("per-page", 50, "items per page")
	groupsSearchCmd.Flags().Int("page", 1, "page number")
	groupsSearchCmd.Flags().Int("per-page", 50, "items per page")

	groupsCmd.AddCommand(groupsListCmd)
	groupsCmd.AddCommand(groupsSearchCmd)
}
