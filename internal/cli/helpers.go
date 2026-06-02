package cli

import (
	"fmt"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

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
