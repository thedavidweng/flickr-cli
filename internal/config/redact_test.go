package config

import "testing"

func TestRedactor(t *testing.T) {
	r := NewRedactor("secret123", "short")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no secrets", "no secrets here", "no secrets here"},
		{"embedded secret", "the secret123 is here", "the *** is here"},
		{"short secret", "short replacement", "*** replacement"},
		{"full value long secret", "secret123", "sec...123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.String(tt.input)
			if got != tt.expected {
				t.Errorf("Redactor.String(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRedactorEmptySecrets(t *testing.T) {
	r := NewRedactor("", "")
	got := r.String("hello world")
	if got != "hello world" {
		t.Errorf("expected no redaction, got %q", got)
	}
}

func TestRedactMap(t *testing.T) {
	m := map[string]any{
		"api_key":        "visible",
		"api_secret":     "hidden",
		"oauth_token":    "visible",
		"password":       "hidden",
		"some_other_key": "visible",
	}

	result := RedactMap(m)

	if result["api_key"] != "visible" {
		t.Error("api_key should not be redacted")
	}
	if result["api_secret"] != "***" {
		t.Error("api_secret should be redacted")
	}
	if result["password"] != "***" {
		t.Error("password should be redacted")
	}
	if result["some_other_key"] != "visible" {
		t.Error("some_other_key should not be redacted")
	}
}
