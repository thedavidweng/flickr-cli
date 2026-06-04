package cli

import (
	"encoding/json"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Direct Flickr API access",
}

var apiCallCmd = &cobra.Command{
	Use:   "call [method]",
	Short: "Call a Flickr API method",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "api.call",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		method := args[0]
		params, _ := cmd.Flags().GetStringToString("param")
		raw, _ := cmd.Flags().GetBool("raw")
		authMode, _ := cmd.Flags().GetString("auth")

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		if authMode == "required" {
			if err := requireAuth(&r, meta, client); err != nil {
				return err
			}
		}

		if authMode == "none" {
			client.OAuthToken = ""
			client.OAuthSecret = ""
		}

		result, err := client.CallRaw(cmd.Context(), method, params)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		if raw {
			var rawJSON any
			json.Unmarshal(result, &rawJSON)
			return r.Success(meta, map[string]any{"raw": rawJSON}, nil)
		}

		var resp any
		json.Unmarshal(result, &resp)
		return r.Success(meta, map[string]any{"response": resp}, nil)
	},
}

var apiMethodsCmd = &cobra.Command{
	Use:   "methods",
	Short: "List available API methods",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "api.methods",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		result, err := client.GetMethods(cmd.Context())
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var resp any
		json.Unmarshal(result, &resp)
		return r.Success(meta, resp, nil)
	},
}

var apiMethodInfoCmd = &cobra.Command{
	Use:   "method-info [method]",
	Short: "Show method parameters and documentation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "api.method-info",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		result, err := client.GetMethodInfo(cmd.Context(), args[0])
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var resp any
		json.Unmarshal(result, &resp)
		return r.Success(meta, resp, nil)
	},
}

func init() {
	apiCallCmd.Flags().StringToString("param", nil, "method parameter key=value (repeatable)")
	apiCallCmd.Flags().Bool("raw", false, "output raw Flickr JSON inside data.raw")
	apiCallCmd.Flags().String("auth", "optional", "auth requirement: optional|required|none")

	apiCmd.AddCommand(apiCallCmd)
	apiCmd.AddCommand(apiMethodsCmd)
	apiCmd.AddCommand(apiMethodInfoCmd)
}
