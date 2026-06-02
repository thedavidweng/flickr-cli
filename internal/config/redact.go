package config

import "strings"

// SensitiveFields lists config fields that should be redacted in output.
var SensitiveFields = []string{
	"api_secret",
	"oauth_token_secret",
	"access_secret",
	"request_secret",
	"consumer_secret",
	"password",
	"mysql_password",
	"dsn",
}

// Redactor redacts sensitive values from strings.
type Redactor struct {
	secrets []string
}

// NewRedactor creates a Redactor that will redact the given secret values.
// Empty values are ignored.
func NewRedactor(values ...string) Redactor {
	var secrets []string
	for _, v := range values {
		if v != "" {
			secrets = append(secrets, v)
		}
	}
	return Redactor{secrets: secrets}
}

// String redacts occurrences of known secrets in s.
// Secrets shorter than 6 chars are fully replaced with "***".
// Longer secrets show first 3 + "..." + last 3 only when they appear
// as a complete field value; embedded occurrences are replaced with "***".
func (r Redactor) String(s string) string {
	result := s
	for _, secret := range r.secrets {
		if len(secret) < 6 {
			result = strings.ReplaceAll(result, secret, "***")
		} else {
			masked := secret[:3] + "..." + secret[len(secret)-3:]
			// If the entire string is the secret, show partial.
			if result == secret {
				result = masked
			} else {
				// Embedded occurrence — full redaction.
				result = strings.ReplaceAll(result, secret, "***")
			}
		}
	}
	return result
}

// RedactMap returns a copy of m with sensitive field values redacted.
// Keys containing "secret", "password", "token_secret", "api_secret", or "dsn" are sensitive.
func RedactMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			result[k] = "***"
		} else {
			result[k] = v
		}
	}
	return result
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	sensitive := []string{"secret", "password", "token_secret", "api_secret", "dsn"}
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
