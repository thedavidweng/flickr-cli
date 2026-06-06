package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// doctorCheck represents a single diagnostic check result.
type doctorCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check configuration and connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "doctor",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		checks := doctorRun(cmd.Context(), app)

		if app.JSON {
			return r.Success(meta, map[string]any{"checks": checks}, nil)
		}

		// Human-readable output
		allOK := true
		for _, c := range checks {
			status := "PASS"
			if !c.OK {
				status = "FAIL"
				allOK = false
			}
			if c.Message != "" {
				r.Human("[%s] %s: %s\n", status, c.Name, c.Message)
			} else {
				r.Human("[%s] %s\n", status, c.Name)
			}
		}
		if allOK {
			r.Human("\nAll checks passed.\n")
		} else {
			r.Human("\nSome checks failed.\n")
		}
		return nil
	},
}

// doctorRun performs all diagnostic checks and returns the results.
func doctorRun(ctx context.Context, app *AppContext) []doctorCheck {
	var checks []doctorCheck

	// 1. Load config file
	cfgPath := app.ConfigFile
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		checks = append(checks, doctorCheck{
			Name:    "config",
			OK:      false,
			Message: err.Error() + "; run 'flickr auth login' to create a profile",
		})
		return checks
	}
	checks = append(checks, doctorCheck{
		Name: "config",
		OK:   true,
	})

	// 2. Check if profile exists
	profile, err := cfg.GetProfile(app.Profile)
	if err != nil {
		checks = append(checks, doctorCheck{
			Name:    "profile",
			OK:      false,
			Message: err.Error() + "; run 'flickr auth login' to create a profile",
		})
		return checks
	}
	checks = append(checks, doctorCheck{
		Name: "profile",
		OK:   true,
	})

	// 3. Check if API key is configured
	creds := config.CredentialsFromProfileAndEnv(profile)
	if !creds.HasAPIKey() {
		checks = append(checks, doctorCheck{
			Name:    "api_key",
			OK:      false,
			Message: "API key is not configured",
		})
		return checks
	}
	checks = append(checks, doctorCheck{
		Name: "api_key",
		OK:   true,
	})

	// 4. Check if OAuth credentials are present
	if !creds.IsAuthenticated() {
		checks = append(checks, doctorCheck{
			Name:    "oauth",
			OK:      false,
			Message: "OAuth credentials are not configured; run 'flickr auth login'",
		})
		return checks
	}
	checks = append(checks, doctorCheck{
		Name: "oauth",
		OK:   true,
	})

	// 5. Test API connection
	client := flickr.NewClient(creds.APIKey, creds.APISecret, creds.OAuthToken, creds.OAuthTokenSecret)
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

	if err := client.TestEcho(ctx); err != nil {
		checks = append(checks, doctorCheck{
			Name:    "api_connection",
			OK:      false,
			Message: err.Error(),
		})
		return checks
	}
	checks = append(checks, doctorCheck{
		Name: "api_connection",
		OK:   true,
	})

	return checks
}
