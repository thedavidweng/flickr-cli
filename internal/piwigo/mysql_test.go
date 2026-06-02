package piwigo

import (
	"context"
	"testing"
)

func TestTablePrefixValidation(t *testing.T) {
	tests := []struct {
		prefix  string
		wantErr bool
	}{
		{"", false},
		{"piwigo_", false},
		{"myTable123", false},
		{"bad-prefix!", true},
		{"bad space", true},
		{"bad.dot", true},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			cfg := DBConfig{DB: "test", User: "root", TablePrefix: tt.prefix}
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBConfigValidate(t *testing.T) {
	// Missing DB
	cfg := DBConfig{User: "root"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing DB")
	}

	// Missing User
	cfg = DBConfig{DB: "test"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing user")
	}

	// Valid config
	cfg = DBConfig{DB: "test", User: "root", TablePrefix: "pwg_"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDSN(t *testing.T) {
	cfg := DBConfig{
		Host:     "localhost",
		Port:     3306,
		DB:       "piwigo",
		User:     "root",
		Password: "secret",
	}

	dsn := DSN(cfg)
	if dsn == "" {
		t.Error("expected non-empty DSN")
	}
}

func TestDSNDefaultPort(t *testing.T) {
	cfg := DBConfig{
		Host: "localhost",
		DB:   "piwigo",
		User: "root",
	}

	dsn := DSN(cfg)
	if dsn == "" {
		t.Error("expected non-empty DSN")
	}
}

func TestOpenInvalidConfig(t *testing.T) {
	// Test with invalid config (missing DB)
	cfg := DBConfig{User: "root"}
	_, err := Open(nil, cfg)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestOpenConnectionError(t *testing.T) {
	// Test with valid config but unreachable host
	cfg := DBConfig{
		Host: "unreachable-host",
		Port: 3306,
		DB:   "test",
		User: "root",
	}
	_, err := Open(context.Background(), cfg)
	if err == nil {
		t.Error("expected error for unreachable host")
	}
}
