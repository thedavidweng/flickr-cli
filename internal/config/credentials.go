package config

import "os"

// Credentials holds the resolved API and OAuth credentials.
type Credentials struct {
	APIKey           string
	APISecret        string
	OAuthToken       string
	OAuthTokenSecret string
}

// CredentialsFromProfileAndEnv resolves credentials from profile and environment overrides.
// Environment variables take precedence over profile values.
func CredentialsFromProfileAndEnv(p *Profile) Credentials {
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

	return creds
}

// IsAuthenticated returns true if the credentials have an OAuth token.
func (c Credentials) IsAuthenticated() bool {
	return c.OAuthToken != "" && c.OAuthTokenSecret != ""
}

// HasAPIKey returns true if the credentials have an API key.
func (c Credentials) HasAPIKey() bool {
	return c.APIKey != ""
}
