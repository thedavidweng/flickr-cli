package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	CurrentProfile string              `yaml:"current_profile"`
	Profiles       map[string]*Profile `yaml:"profiles"`
	State          ConfigState         `yaml:"state,omitempty"`
}

// ConfigState holds transient CLI state.
type ConfigState struct {
	PendingOAuth map[string]PendingOAuthEntry `yaml:"pending_oauth,omitempty"`
}

// PendingOAuthEntry holds state for an in-flight OAuth begin/complete flow.
type PendingOAuthEntry struct {
	Profile       string `yaml:"profile"`
	RequestToken  string `yaml:"request_token"`
	RequestSecret string `yaml:"request_secret"`
	Perms         string `yaml:"perms"`
	ExpiresAt     string `yaml:"expires_at"`
}

// UserInfo holds the authenticated user's identity.
type UserInfo struct {
	NSID     string `yaml:"nsid"`
	Username string `yaml:"username"`
	Fullname string `yaml:"fullname,omitempty"`
}

// Profile represents a named configuration profile.
type Profile struct {
	APIKey           string       `yaml:"api_key"`
	APISecret        string       `yaml:"api_secret"`
	OAuthToken       string       `yaml:"oauth_token"`
	OAuthTokenSecret string       `yaml:"oauth_token_secret"`
	Permissions      string       `yaml:"permissions"`
	User             UserInfo     `yaml:"user"`
	CreatedAt        string       `yaml:"created_at,omitempty"`
	UpdatedAt        string       `yaml:"updated_at,omitempty"`
	CachePath        string       `yaml:"cache_path"`
	AuditLogPath     string       `yaml:"audit_log_path"`
	Backup           BackupConfig `yaml:"backup"`
	Upload           UploadConfig `yaml:"upload"`
	Endpoints        Endpoints    `yaml:"endpoints,omitempty"`
}

// Endpoints holds API endpoint overrides (for testing or self-hosted instances).
type Endpoints struct {
	REST         string `yaml:"rest,omitempty"`
	Upload       string `yaml:"upload,omitempty"`
	RequestToken string `yaml:"request_token,omitempty"`
	Authorize    string `yaml:"authorize,omitempty"`
	AccessToken  string `yaml:"access_token,omitempty"`
}

// BackupConfig holds backup-related defaults.
type BackupConfig struct {
	Dest     string `yaml:"dest"`
	Metadata string `yaml:"metadata"`
	Resume   bool   `yaml:"resume"`
}

// UploadConfig holds upload-related defaults.
type UploadConfig struct {
	Dedupe string `yaml:"dedupe"`
	Hash   string `yaml:"hash"`
}

// Load reads and parses the config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
	}
	return &cfg, nil
}

// Save writes the config to path with secure permissions.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to temp file, then atomic rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming config: %w", err)
	}
	return nil
}

// LoadOrCreate loads the config or creates a default one if it doesn't exist.
func LoadOrCreate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, nil
	}
	cfg = &Config{
		CurrentProfile: "default",
		Profiles:       map[string]*Profile{},
	}
	if err := Save(path, cfg); err != nil {
		return nil, fmt.Errorf("creating default config: %w", err)
	}
	return cfg, nil
}

// GetProfile returns the named profile or an error if not found.
func (c *Config) GetProfile(name string) (*Profile, error) {
	p, ok := c.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

// SetProfile sets or updates a named profile.
func (c *Config) SetProfile(name string, p *Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	c.Profiles[name] = p
}

// SetPendingOAuth stores a pending OAuth entry under the given ID.
func (c *Config) SetPendingOAuth(id string, entry *PendingOAuthEntry) {
	if c.State.PendingOAuth == nil {
		c.State.PendingOAuth = make(map[string]PendingOAuthEntry)
	}
	c.State.PendingOAuth[id] = *entry
}

// GetPendingOAuth retrieves a pending OAuth entry by ID.
func (c *Config) GetPendingOAuth(id string) (PendingOAuthEntry, bool) {
	entry, ok := c.State.PendingOAuth[id]
	return entry, ok
}

// DeletePendingOAuth removes a pending OAuth entry by ID.
func (c *Config) DeletePendingOAuth(id string) {
	delete(c.State.PendingOAuth, id)
}

// PurgeExpiredPendingOAuth removes all expired pending OAuth entries.
func (c *Config) PurgeExpiredPendingOAuth() {
	now := time.Now().UTC()
	for id, entry := range c.State.PendingOAuth {
		expires, err := time.Parse(time.RFC3339, entry.ExpiresAt)
		if err == nil && now.After(expires) {
			delete(c.State.PendingOAuth, id)
		}
	}
}
