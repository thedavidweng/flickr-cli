package config

import (
	"os"
	"testing"
)

func TestCredentialsFromProfileAndEnv(t *testing.T) {
	p := &Profile{
		APIKey:           "profile-key",
		APISecret:        "profile-secret",
		OAuthToken:       "profile-token",
		OAuthTokenSecret: "profile-token-secret",
	}

	creds := CredentialsFromProfileAndEnv(p)
	if creds.APIKey != "profile-key" {
		t.Errorf("expected profile-key, got %s", creds.APIKey)
	}
	if creds.APISecret != "profile-secret" {
		t.Errorf("expected profile-secret, got %s", creds.APISecret)
	}
	if creds.OAuthToken != "profile-token" {
		t.Errorf("expected profile-token, got %s", creds.OAuthToken)
	}
	if creds.OAuthTokenSecret != "profile-token-secret" {
		t.Errorf("expected profile-token-secret, got %s", creds.OAuthTokenSecret)
	}
}

func TestCredentialsEnvOverride(t *testing.T) {
	_ = os.Setenv("FLICKR_API_KEY", "env-key")
	_ = os.Setenv("FLICKR_API_SECRET", "env-secret")
	_ = os.Setenv("FLICKR_OAUTH_TOKEN", "env-token")
	_ = os.Setenv("FLICKR_OAUTH_TOKEN_SECRET", "env-token-secret")
	defer func() { _ = os.Unsetenv("FLICKR_API_KEY") }()
	defer func() { _ = os.Unsetenv("FLICKR_API_SECRET") }()
	defer func() { _ = os.Unsetenv("FLICKR_OAUTH_TOKEN") }()
	defer func() { _ = os.Unsetenv("FLICKR_OAUTH_TOKEN_SECRET") }()

	p := &Profile{
		APIKey:    "profile-key",
		APISecret: "profile-secret",
	}

	creds := CredentialsFromProfileAndEnv(p)
	if creds.APIKey != "env-key" {
		t.Errorf("expected env-key, got %s", creds.APIKey)
	}
	if creds.APISecret != "env-secret" {
		t.Errorf("expected env-secret, got %s", creds.APISecret)
	}
	if creds.OAuthToken != "env-token" {
		t.Errorf("expected env-token, got %s", creds.OAuthToken)
	}
	if creds.OAuthTokenSecret != "env-token-secret" {
		t.Errorf("expected env-token-secret, got %s", creds.OAuthTokenSecret)
	}
}

func TestCredentialsIsAuthenticated(t *testing.T) {
	creds := Credentials{OAuthToken: "tok", OAuthTokenSecret: "sec"}
	if !creds.IsAuthenticated() {
		t.Error("expected authenticated")
	}
	creds2 := Credentials{OAuthToken: "tok"}
	if creds2.IsAuthenticated() {
		t.Error("expected not authenticated without secret")
	}
	creds3 := Credentials{}
	if creds3.IsAuthenticated() {
		t.Error("expected not authenticated without token")
	}
}

func TestCredentialsHasAPIKey(t *testing.T) {
	creds := Credentials{APIKey: "key"}
	if !creds.HasAPIKey() {
		t.Error("expected HasAPIKey=true")
	}
	creds2 := Credentials{}
	if creds2.HasAPIKey() {
		t.Error("expected HasAPIKey=false")
	}
}
