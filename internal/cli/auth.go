package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Flickr authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Create or refresh OAuth credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
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
		creds := config.CredentialsFromProfileAndEnv(profile)
		if creds.IsAuthenticated() && !force {
			client := flickr.NewClient(creds.APIKey, creds.APISecret, creds.OAuthToken, creds.OAuthTokenSecret)
			loginInfo, err := client.TestLogin(cmd.Context())
			if err != nil {
				return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "test login failed: %v", err))
			}
			r.Human("Already authenticated\n")
			return r.Success(meta, loginInfo, nil)
		}

		// Get API key
		apiKey := creds.APIKey
		if apiKey == "" {
			apiKeyFlag, _ := cmd.Flags().GetString("api-key")
			if apiKeyFlag != "" {
				apiKey = apiKeyFlag
			} else if !isTerminal() {
				return r.Failure(meta, output.Errorf(model.ErrConfig, "API key required in non-interactive mode"))
			} else {
				fmt.Fprintln(r.Err, "Get your API key and secret at: https://www.flickr.com/services/apps/")
				fmt.Fprintln(r.Err)
				fmt.Fprint(r.Err, "API key: ")
				apiKey = readLine()
			}
		}

		// Get API secret
		apiSecret := creds.APISecret
		if apiSecret == "" {
			apiSecretFlag, _ := cmd.Flags().GetString("api-secret")
			if apiSecretFlag != "" {
				apiSecret = apiSecretFlag
			} else if envName, _ := cmd.Flags().GetString("api-secret-env"); envName != "" {
				apiSecret = os.Getenv(envName)
			} else if !isTerminal() {
				return r.Failure(meta, output.Errorf(model.ErrConfig, "API secret required in non-interactive mode"))
			} else {
				fmt.Fprint(r.Err, "API secret: ")
				apiSecret = readLine()
			}
		}

		if apiKey == "" || apiSecret == "" {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "API key and secret are required"))
		}

		// OAuth flow
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

		perms, _ := cmd.Flags().GetString("perms")
		callbackType, _ := cmd.Flags().GetString("callback")

		var callbackURL string
		if callbackType == "oob" {
			callbackURL = "oob"
		} else {
			callbackURL = "http://localhost"
		}

		// Get request token
		reqToken, err := client.RequestToken(cmd.Context(), callbackURL)
		if err != nil {
			if strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") {
				return r.Failure(meta, output.ErrorWithDetails(
					model.ErrAuthFailed,
					"Invalid API key or secret. Verify your credentials at https://www.flickr.com/services/apps/",
					map[string]any{"profile": app.Profile},
				))
			}
			return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "requesting token: %v", err))
		}

		var verifier string
		if callbackType == "oob" {
			// OOB flow: print URL and prompt for verifier
			authURL := client.AuthorizationURL(reqToken.Token, perms)
			r.Human("Open this URL to authorize:\n%s\n\n", authURL)
			if !isTerminal() {
				return r.Failure(meta, output.Errorf(model.ErrConfig, "verifier required in non-interactive mode"))
			}
			fmt.Fprint(r.Err, "Enter verifier code: ")
			verifier = readLine()
		} else {
			// Localhost callback flow
			port, _ := cmd.Flags().GetInt("callback-port")
			verifier, err = localhostCallback(cmd.Context(), client, reqToken.Token, perms, port)
			if err != nil {
				return r.Failure(meta, output.Errorf(model.ErrAuthFailed, "callback: %v", err))
			}
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

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
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

		creds := config.CredentialsFromProfileAndEnv(profile)
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
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}
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

		// Clear OAuth credentials
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

func localhostCallback(ctx context.Context, client *flickr.Client, token, perms string, port int) (string, error) {
	authURL := client.AuthorizationURL(token, perms)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return "", fmt.Errorf("listening on port %d: %w", port, err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	redirectURI := fmt.Sprintf("http://localhost:%d", addr.Port)

	fmt.Fprintf(os.Stderr, "Open this URL to authorize:\n%s\n", authURL)
	fmt.Fprintf(os.Stderr, "Waiting for callback on %s...\n", redirectURI)

	verifierCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query().Get("oauth_verifier")
		if v == "" {
			http.Error(w, "missing verifier", http.StatusBadRequest)
			errCh <- fmt.Errorf("missing verifier in callback")
			return
		}
		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		verifierCh <- v
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			// Server closed is expected
		}
	}()

	select {
	case verifier := <-verifierCh:
		srv.Shutdown(ctx)
		return verifier, nil
	case err := <-errCh:
		srv.Shutdown(ctx)
		return "", err
	case <-ctx.Done():
		srv.Shutdown(ctx)
		return "", ctx.Err()
	}
}
