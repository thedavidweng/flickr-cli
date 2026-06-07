package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// newRenderer creates a Renderer from AppContext and cobra.Command,
// propagating all output-related flags (Quiet, NoColor, etc.).
func newRenderer(app *AppContext, cmd *cobra.Command) output.Renderer {
	return output.Renderer{
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
		JSON:    app.JSON,
		Pretty:  app.Pretty,
		Compact: app.Compact,
		Full:    app.Full,
		Quiet:   app.Quiet,
		NoColor: app.NoColor,
		Verbose: app.Verbose,
	}
}

// CmdContext bundles everything a command handler needs.
type CmdContext struct {
	App    *AppContext
	Cmd    *cobra.Command
	Args   []string
	Client *flickr.Client
	Config *config.Config
	R      output.Renderer
	Meta   output.RuntimeMetaInput
}

// CmdFunc is a command handler that receives a ready-to-use context.
type CmdFunc func(ctx *CmdContext) error

// withAuth wraps a CmdFunc: loads config, creates client, checks auth.
func withAuth(command string, fn CmdFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   command,
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, cfg, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}
		return fn(&CmdContext{App: app, Cmd: cmd, Args: args, Client: client, Config: cfg, R: r, Meta: meta})
	}
}

// getClient creates a Flickr client from the current app context and config.
func getClient(app *AppContext) (*flickr.Client, *config.Config, error) {
	cfgPath := app.ConfigFile
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("not configured. Run 'flickr auth login' to get started")
	}

	profile, err := cfg.GetProfile(app.Profile)
	if err != nil {
		return nil, cfg, fmt.Errorf("not authenticated. Run 'flickr auth login' to get started")
	}

	creds := config.CredentialsFromProfileAndEnv(profile)
	client := flickr.NewClient(creds.APIKey, creds.APISecret, creds.OAuthToken, creds.OAuthTokenSecret)
	client.Retries = app.Retries
	client.RequestInterval = app.RequestInterval

	// Apply endpoint overrides from config (used for testing)
	if profile.Endpoints.REST != "" {
		client.Endpoints.REST = profile.Endpoints.REST
	}
	if profile.Endpoints.Upload != "" {
		client.Endpoints.Upload = profile.Endpoints.Upload
	}
	if profile.Endpoints.RequestToken != "" {
		client.Endpoints.RequestToken = profile.Endpoints.RequestToken
	}
	if profile.Endpoints.Authorize != "" {
		client.Endpoints.Authorize = profile.Endpoints.Authorize
	}
	if profile.Endpoints.AccessToken != "" {
		client.Endpoints.AccessToken = profile.Endpoints.AccessToken
	}

	return client, cfg, nil
}

// requireAuth checks that the client is authenticated.
func requireAuth(r *output.Renderer, meta output.RuntimeMetaInput, client *flickr.Client) error {
	if !client.IsAuthenticated() {
		return r.Failure(meta, output.ErrorWithDetails(
			model.ErrAuthRequired,
			"Authentication required. Run 'flickr auth login' to authenticate.",
			map[string]any{"profile": meta.Profile},
		))
	}
	return nil
}
