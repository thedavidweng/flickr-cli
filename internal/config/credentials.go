package config

import (
	"fmt"
	"os"
	"strings"
)

// Credentials holds the resolved API and OAuth credentials.
type Credentials struct {
	APIKey           string
	APISecret        string
	OAuthToken       string
	OAuthTokenSecret string
}

// CredentialsFromProfileAndEnv resolves credentials from profile and environment overrides.
// Environment variables take precedence over profile values. Any resolved value of the
// form "env:NAME" is an indirection: it is replaced with the value of $NAME, and it is an
// error if $NAME is unset. Indirection keeps secrets out of the 0600 config file.
func CredentialsFromProfileAndEnv(p *Profile) (Credentials, error) {
	creds := Credentials{
		APIKey:           p.APIKey,
		APISecret:        p.APISecret,
		OAuthToken:       p.OAuthToken,
		OAuthTokenSecret: p.OAuthTokenSecret,
	}

	if v := os.Getenv("FLICKR_API_KEY"); v != "" {
		creds.APIKey = v
	}
	if v := os.Getenv("FLICKR_API_SECRET"); v != "" {
		creds.APISecret = v
	}
	if v := os.Getenv("FLICKR_OAUTH_TOKEN"); v != "" {
		creds.OAuthToken = v
	}
	if v := os.Getenv("FLICKR_OAUTH_TOKEN_SECRET"); v != "" {
		creds.OAuthTokenSecret = v
	}

	var err error
	if creds.APIKey, err = resolveEnvIndirection("api_key", creds.APIKey); err != nil {
		return Credentials{}, err
	}
	if creds.APISecret, err = resolveEnvIndirection("api_secret", creds.APISecret); err != nil {
		return Credentials{}, err
	}
	if creds.OAuthToken, err = resolveEnvIndirection("oauth_token", creds.OAuthToken); err != nil {
		return Credentials{}, err
	}
	if creds.OAuthTokenSecret, err = resolveEnvIndirection("oauth_token_secret", creds.OAuthTokenSecret); err != nil {
		return Credentials{}, err
	}

	return creds, nil
}

// resolveEnvIndirection expands an "env:NAME" indirection to the value of $NAME.
// Non-indirection values pass through unchanged. An unset $NAME is an error.
func resolveEnvIndirection(field, value string) (string, error) {
	const prefix = "env:"
	if !strings.HasPrefix(value, prefix) {
		return value, nil
	}
	name := strings.TrimPrefix(value, prefix)
	resolved := os.Getenv(name)
	if resolved == "" {
		return "", fmt.Errorf("%s references env var %q which is not set", field, name)
	}
	return resolved, nil
}

// IsAuthenticated returns true if the credentials have an OAuth token.
func (c Credentials) IsAuthenticated() bool {
	return c.OAuthToken != "" && c.OAuthTokenSecret != ""
}

// HasAPIKey returns true if the credentials have an API key.
func (c Credentials) HasAPIKey() bool {
	return c.APIKey != ""
}
