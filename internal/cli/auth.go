package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// Package-level seams for testing.
var (
	readInput = func() string { return readFrom(os.Stdin) }
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Flickr authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Create or refresh OAuth credentials",
	Long: `Authenticate with Flickr using OAuth 1.0a.

A Flickr API key and secret are required. Get yours at:
  https://www.flickr.com/services/apps/

Note: Flickr requires a Pro account to register an API application.
Free accounts cannot create API keys as of 2024.

  flickr auth login                              Interactive prompt for key/secret
  flickr auth login --api-key KEY --api-secret S Non-interactive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "auth.login",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		// Load config
		cfgPath := app.ConfigFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.LoadOrCreate(cfgPath)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "loading config: %v", err))
		}

		profile, _ := cfg.GetProfile(app.Profile)
		if profile == nil {
			profile = &config.Profile{}
		}

		// Check if already authenticated (unless --force)
		force, _ := cmd.Flags().GetBool("force")
		creds, err := config.CredentialsFromProfileAndEnv(profile)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if creds.IsAuthenticated() && !force {
			client := flickr.NewClient(creds.APIKey, creds.APISecret, creds.OAuthToken, creds.OAuthTokenSecret)
			loginInfo, err := client.TestLogin(cmd.Context())
			if err != nil {
				return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "test login failed: %v", err))
			}
			r.Human("Already authenticated\n")
			return r.Success(meta, loginInfo, nil)
		}

		// Resolve API key and secret
		apiKey, apiSecret, err := resolveCredentials(cmd, &r, meta, creds)
		if err != nil {
			return err
		}

		// OAuth flow
		callbackType, _ := cmd.Flags().GetString("callback")
		perms, _ := cmd.Flags().GetString("perms")

		client := flickr.NewClient(apiKey, apiSecret, "", "")

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

		if callbackType == "" {
			if isTerminal() {
				callbackType = "localhost"
			} else {
				callbackType = "oob"
			}
		}

		var verifier string
		var reqToken *flickr.RequestTokenResponse

		flow := &flickr.OAuthFlow{Client: client}

		if callbackType == "oob" {
			authURL, rt, err := flow.OOBRequestToken(cmd.Context(), perms)
			if err != nil {
				return handleRequestTokenError(r, meta, err)
			}
			reqToken = rt
			verifier, err = oobAuthorize(authURL)
			if err != nil {
				return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "authorization: %v", err))
			}
		} else {
			port, _ := cmd.Flags().GetInt("callback-port")
			rt, v, err := flow.LocalhostFlow(cmd.Context(), perms, port, func(authURL, callbackURL string) {
				_, _ = fmt.Fprintln(os.Stderr)
				_, _ = fmt.Fprintf(os.Stderr, "Open this URL to authorize:\n\n  %s\n\n", authURL)
				if callbackURL != "" && !strings.HasPrefix(callbackURL, "http://localhost") && !strings.HasPrefix(callbackURL, "http://127.0.0.1") {
					_, _ = fmt.Fprintf(os.Stderr, "Callback will be received on %s\n", callbackURL)
				}
			})
			if err != nil {
				return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "authorization: %v", err))
			}
			reqToken = rt
			verifier = v
		}

		// Exchange for access token
		accessToken, err := client.AccessToken(cmd.Context(), reqToken.Token, reqToken.TokenSecret, verifier)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "access token: %v", err))
		}

		// Save credentials
		profile.APIKey = apiKey
		profile.APISecret = apiSecret
		profile.OAuthToken = accessToken.Token
		profile.OAuthTokenSecret = accessToken.TokenSecret
		profile.User.NSID = accessToken.UserNSID
		profile.User.Username = accessToken.Username
		profile.Permissions = perms
		cfg.SetProfile(app.Profile, profile)

		if err := config.Save(cfgPath, cfg); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "saving config: %v", err))
		}

		r.Human("Authenticated as %s (%s)\n", accessToken.Username, accessToken.UserNSID)
		return r.Success(meta, map[string]any{
			"user_id":     accessToken.UserNSID,
			"username":    accessToken.Username,
			"permissions": perms,
		}, nil)
	},
}

func handleRequestTokenError(r output.Renderer, meta output.RuntimeMetaInput, err error) error {
	if strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") {
		return r.Failure(meta, output.ErrorWithDetails(
			model.ErrAuthFailed,
			"Invalid API key or secret. Verify your credentials at https://www.flickr.com/services/apps/",
			nil,
		))
	}
	return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "requesting token: %v", err))
}

// resolveCredentials reads the API key and secret from flags, environment, or interactive prompt.
func resolveCredentials(cmd *cobra.Command, r *output.Renderer, meta output.RuntimeMetaInput, creds config.Credentials) (apiKey, apiSecret string, err error) {
	apiKey = creds.APIKey
	if apiKey == "" {
		switch {
		case cmd.Flags().Changed("api-key"):
			apiKey, _ = cmd.Flags().GetString("api-key")
		case !isTerminal():
			return "", "", r.Failure(meta, output.Errorf(model.ErrConfig, "API key required. Get one at https://www.flickr.com/services/apps/"))
		default:
			_, _ = fmt.Fprintln(r.Err, "A Flickr API key and secret are required.")
			_, _ = fmt.Fprintln(r.Err, "Get yours at: https://www.flickr.com/services/apps/")
			_, _ = fmt.Fprintln(r.Err)
			_, _ = fmt.Fprint(r.Err, "API key: ")
			apiKey = readInput()
		}
	}

	apiSecret = creds.APISecret
	if apiSecret == "" {
		switch {
		case cmd.Flags().Changed("api-secret"):
			apiSecret, _ = cmd.Flags().GetString("api-secret")
		case cmd.Flags().Changed("api-secret-env"):
			envName, _ := cmd.Flags().GetString("api-secret-env")
			apiSecret = os.Getenv(envName)
		case !isTerminal():
			return "", "", r.Failure(meta, output.Errorf(model.ErrConfig, "API secret required. Get one at https://www.flickr.com/services/apps/"))
		default:
			_, _ = fmt.Fprint(r.Err, "API secret: ")
			apiSecret = readInput()
		}
	}

	if apiKey == "" || apiSecret == "" {
		return "", "", r.Failure(meta, output.Errorf(model.ErrConfig, "API key and secret are required"))
	}

	return apiKey, apiSecret, nil
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "auth.status",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		cfgPath := app.ConfigFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrAuthRequired,
				"Not configured. Run 'flickr auth login' to get started.",
				map[string]any{"profile": app.Profile},
			))
		}

		profile, _ := cfg.GetProfile(app.Profile)
		if profile == nil {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrAuthRequired,
				fmt.Sprintf("Profile %q not configured. Run 'flickr auth login' to get started.", app.Profile),
				map[string]any{"profile": app.Profile},
			))
		}

		creds, err := config.CredentialsFromProfileAndEnv(profile)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if !creds.HasAPIKey() {
			return r.Failure(meta, output.Errorf(model.ErrAuthRequired, "no API key configured"))
		}
		if !creds.IsAuthenticated() {
			return r.Failure(meta, output.Errorf(model.ErrAuthRequired, "not authenticated"))
		}

		client := flickr.NewClient(creds.APIKey, creds.APISecret, creds.OAuthToken, creds.OAuthTokenSecret)
		loginInfo, err := client.TestLogin(cmd.Context())
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "test login failed: %v", err))
		}

		r.Human("Authenticated\n")
		return r.Success(meta, map[string]any{
			"authenticated": true,
			"user_id":       profile.User.NSID,
			"username":      profile.User.Username,
			"permissions":   profile.Permissions,
			"login_info":    loginInfo,
		}, nil)
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials for current profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "auth.logout",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		cfgPath := app.ConfigFile
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "loading config: %v", err))
		}

		profile, _ := cfg.GetProfile(app.Profile)
		if profile == nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "no profile %q configured", app.Profile))
		}

		if app.DryRun {
			r.Human("Would clear OAuth credentials for profile %q\n", app.Profile)
			return r.Success(meta, map[string]any{
				"planned": true,
				"profile": app.Profile,
				"action":  "clear_credentials",
				"fields":  []string{"oauth_token", "oauth_token_secret", "permissions"},
			}, nil)
		}

		// Clear OAuth credentials but keep API key (user may want to re-auth)
		profile.OAuthToken = ""
		profile.OAuthTokenSecret = ""
		profile.Permissions = ""
		cfg.SetProfile(app.Profile, profile)

		if err := config.Save(cfgPath, cfg); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "saving config: %v", err))
		}

		r.Human("Logged out from profile %q\n", app.Profile)
		return r.Success(meta, map[string]any{
			"profile": app.Profile,
			"action":  "cleared_credentials",
		}, nil)
	},
}

func init() {
	authLoginCmd.Flags().String("perms", "read", "authorization permission: read|write|delete")
	authLoginCmd.Flags().Bool("force", false, "force re-authentication")
	authLoginCmd.Flags().String("callback", "localhost", "auth callback strategy: localhost|oob")
	authLoginCmd.Flags().Int("callback-port", 0, "local callback port, 0 for auto")
	authLoginCmd.Flags().String("api-key", "", "Flickr API key")
	authLoginCmd.Flags().String("api-secret", "", "Flickr API secret")
	authLoginCmd.Flags().String("api-secret-env", "", "env var name containing API secret")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func readLine() string {
	return readFrom(os.Stdin)
}

func readFrom(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// oobAuthorize handles the out-of-band authorization user interaction.
// It displays the auth URL and reads the verifier code from stdin.
func oobAuthorize(authURL string) (string, error) {
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprintln(os.Stderr, "Open this URL to authorize:")
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprintf(os.Stderr, "  %s\n", authURL)
	_, _ = fmt.Fprintln(os.Stderr)
	_, _ = fmt.Fprintln(os.Stderr, "After authorizing, Flickr will show a verification code.")
	_, _ = fmt.Fprintln(os.Stderr, "Paste it here to complete authentication.")
	_, _ = fmt.Fprintln(os.Stderr)

	if !isTerminal() {
		return "", fmt.Errorf("verifier code required in non-interactive mode; use --callback localhost or pipe the code to stdin")
	}

	_, _ = fmt.Fprint(os.Stderr, "Verification code: ")
	verifier := readInput()
	if verifier == "" {
		return "", fmt.Errorf("no verification code entered")
	}

	return verifier, nil
}
