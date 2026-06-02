package config

import "testing"

func TestProfileNames(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			"default": {},
			"work":    {},
		},
	}

	names := cfg.ProfileNames()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestActiveProfile(t *testing.T) {
	cfg := &Config{CurrentProfile: "work"}
	if cfg.ActiveProfile() != "work" {
		t.Errorf("expected work, got %s", cfg.ActiveProfile())
	}

	cfg2 := &Config{}
	if cfg2.ActiveProfile() != "default" {
		t.Errorf("expected default, got %s", cfg2.ActiveProfile())
	}
}
