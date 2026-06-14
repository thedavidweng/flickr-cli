package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := &Config{
		CurrentProfile: "work",
		Profiles: map[string]*Profile{
			"work": {
				APIKey:           "key",
				APISecret:        "secret",
				OAuthToken:       "token",
				OAuthTokenSecret: "token-secret",
				Permissions:      "write",
				User: UserInfo{
					NSID:     "98765@N02",
					Username: "roundtrip",
					Fullname: "Round Trip",
				},
				CreatedAt: "2026-06-02T12:00:00Z",
				UpdatedAt: "2026-06-02T13:00:00Z",
			},
		},
		State: ConfigState{
			PendingOAuth: map[string]PendingOAuthEntry{
				"test-uuid": {
					Profile:       "work",
					RequestToken:  "req-token",
					RequestSecret: "req-secret",
					Perms:         "write",
					ExpiresAt:     "2026-06-02T12:30:00Z",
				},
			},
		},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.CurrentProfile != "work" {
		t.Errorf("current_profile: got %s, want work", loaded.CurrentProfile)
	}

	p := loaded.Profiles["work"]
	if p.User.NSID != "98765@N02" {
		t.Errorf("user.nsid: got %s, want 98765@N02", p.User.NSID)
	}
	if p.User.Fullname != "Round Trip" {
		t.Errorf("user.fullname: got %s, want Round Trip", p.User.Fullname)
	}
	if p.CreatedAt != "2026-06-02T12:00:00Z" {
		t.Errorf("created_at: got %s", p.CreatedAt)
	}

	pending := loaded.State.PendingOAuth["test-uuid"]
	if pending.RequestToken != "req-token" {
		t.Errorf("pending request_token: got %s", pending.RequestToken)
	}
}

func TestPendingOAuthAddAndGet(t *testing.T) {
	cfg := &Config{Profiles: map[string]*Profile{}}

	entry := PendingOAuthEntry{
		Profile:       "default",
		RequestToken:  "req-token",
		RequestSecret: "req-secret",
		Perms:         "write",
		ExpiresAt:     time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339),
	}
	cfg.SetPendingOAuth("abc-123", &entry)

	got, ok := cfg.GetPendingOAuth("abc-123")
	if !ok {
		t.Fatal("expected pending entry to exist")
	}
	if got.RequestToken != "req-token" {
		t.Errorf("request_token: got %s, want req-token", got.RequestToken)
	}
	if got.Profile != "default" {
		t.Errorf("profile: got %s, want default", got.Profile)
	}
}

func TestPendingOAuthDelete(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{},
		State: ConfigState{
			PendingOAuth: map[string]PendingOAuthEntry{
				"to-delete": {
					Profile:      "default",
					RequestToken: "old",
					ExpiresAt:    time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339),
				},
			},
		},
	}

	cfg.DeletePendingOAuth("to-delete")

	_, ok := cfg.GetPendingOAuth("to-delete")
	if ok {
		t.Error("expected pending entry to be deleted")
	}
}

func TestPendingOAuthPurgeExpired(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{},
		State: ConfigState{
			PendingOAuth: map[string]PendingOAuthEntry{
				"expired": {
					Profile:      "default",
					RequestToken: "old",
					ExpiresAt:    time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339),
				},
				"valid": {
					Profile:      "default",
					RequestToken: "new",
					ExpiresAt:    time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339),
				},
			},
		},
	}

	cfg.PurgeExpiredPendingOAuth()

	_, expiredOK := cfg.GetPendingOAuth("expired")
	if expiredOK {
		t.Error("expired entry should have been purged")
	}
	_, validOK := cfg.GetPendingOAuth("valid")
	if !validOK {
		t.Error("valid entry should not have been purged")
	}
}

func TestSensitiveFieldsCoveredByRedactor(t *testing.T) {
	secrets := []string{"api_secret", "oauth_token_secret", "request_secret"}
	for _, field := range secrets {
		found := false
		for _, sf := range SensitiveFields {
			if sf == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SensitiveFields should include %q", field)
		}
	}

	m := map[string]any{
		"api_key":            "visible",
		"api_secret":         "should-be-hidden",
		"oauth_token":        "also-visible",
		"oauth_token_secret": "should-be-hidden-too",
	}
	redacted := RedactMap(m)
	if redacted["api_key"] != "visible" {
		t.Error("api_key should not be redacted")
	}
	if redacted["api_secret"] == "should-be-hidden" {
		t.Error("api_secret should be redacted")
	}
	if redacted["oauth_token_secret"] == "should-be-hidden-too" {
		t.Error("oauth_token_secret should be redacted")
	}
}
