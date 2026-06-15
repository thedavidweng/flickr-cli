package flickr

import (
	"testing"
	"time"
)

func TestPercentEncode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello world", "hello%20world"},
		{"foo/bar", "foo%2Fbar"},
		{"a+b", "a%2Bb"},
		{"abc123", "abc123"},
		{"-._~-._~-", "-._~-._~-"},
		{"", ""},
		{"hello/world/again", "hello%2Fworld%2Fagain"},
		{"key=value", "key%3Dvalue"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := PercentEncode(tt.input)
			if got != tt.expected {
				t.Errorf("PercentEncode(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeParams(t *testing.T) {
	params := map[string][]string{
		"c": {"3"},
		"a": {"1"},
		"b": {"2"},
	}
	got := NormalizeParams(params)
	expected := "a=1&b=2&c=3"
	if got != expected {
		t.Errorf("NormalizeParams = %q, want %q", got, expected)
	}
}

func TestNormalizeParamsDuplicateKeys(t *testing.T) {
	params := map[string][]string{
		"a": {"1", "2"},
	}
	got := NormalizeParams(params)
	expected := "a=1&a=2"
	if got != expected {
		t.Errorf("NormalizeParams = %q, want %q", got, expected)
	}
}

func TestSignatureBaseString(t *testing.T) {
	// Example from OAuth spec
	params := map[string][]string{
		"status": {"Hello Ladies + Gentlemen"},
	}
	got := SignatureBaseString("POST", "https://api.example.com/statuses/update.json", params)
	if got == "" {
		t.Error("expected non-empty signature base string")
	}
	// Verify it starts with POST& and encodes the URL
	if len(got) < 5 || got[:5] != "POST&" {
		t.Errorf("expected to start with POST&, got %q", got[:10])
	}
}

func TestHMACSHA1Signature(t *testing.T) {
	// Test with known values
	baseString := "POST&https%3A%2F%2Fapi.example.com%2Fstatuses%2Fupdate.json&status%3DHello%2520Ladies%2520%252B%2520Gentlemen"
	signingKey := "consumer_secret&token_secret"
	sig := HMACSHA1Signature(baseString, signingKey)
	if sig == "" {
		t.Error("expected non-empty signature")
	}
	// Signature should be base64 encoded
	if len(sig) < 10 {
		t.Errorf("signature too short: %q", sig)
	}
}

func TestSigningKey(t *testing.T) {
	got := SigningKey("cs", "ts")
	expected := "cs&ts"
	if got != expected {
		t.Errorf("SigningKey = %q, want %q", got, expected)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	params := map[string]string{
		"oauth_consumer_key": "key",
		"oauth_token":        "token",
	}
	header := AuthorizationHeader(params)
	if header[:6] != "OAuth " {
		t.Errorf("expected header to start with 'OAuth ', got %q", header[:10])
	}
}

func TestOAuthSignerSign(t *testing.T) {
	signer := NewOAuthSigner(OAuthCredentials{
		ConsumerKey:    "ck",
		ConsumerSecret: "cs",
		Token:          "tk",
		TokenSecret:    "ts",
	})
	// Override Now and Nonce for deterministic tests
	signer.Now = func() time.Time { return time.Unix(1234567890, 0) }
	signer.Nonce = func() string { return "nonce123" }

	params := map[string][]string{
		"status": {"hello"},
	}

	oauthParams, err := signer.Sign("POST", "https://api.example.com/endpoint", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if oauthParams["oauth_consumer_key"] != "ck" {
		t.Errorf("expected consumer key ck, got %s", oauthParams["oauth_consumer_key"])
	}
	if oauthParams["oauth_token"] != "tk" {
		t.Errorf("expected token tk, got %s", oauthParams["oauth_token"])
	}
	if oauthParams["oauth_signature_method"] != "HMAC-SHA1" {
		t.Errorf("expected HMAC-SHA1, got %s", oauthParams["oauth_signature_method"])
	}
	if oauthParams["oauth_signature"] == "" {
		t.Error("expected non-empty signature")
	}
}

func TestOAuthSpecSampleSignature(t *testing.T) {
	// OAuth 1.0a spec example from https://datatracker.ietf.org/doc/html/rfc5849#section-3.4.1
	signer := OAuthSigner{
		Creds: OAuthCredentials{
			ConsumerKey:    "dpf43f3p2l4k3l03",
			ConsumerSecret: "kd94hf93k423kf44",
			Token:          "nnch734d00sl2jdk",
			TokenSecret:    "pfkkdhi9sl3r4s00",
		},
		Now:   func() time.Time { return time.Unix(1191242096, 0) },
		Nonce: func() string { return "kllo9940pd9333jh" },
	}

	params := map[string][]string{}

	oauthParams, err := signer.Sign("POST", "http://photos.example.net/request_token", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "fG78XbuIg3Ga75N395fQzXvgRXM="
	if oauthParams["oauth_signature"] != expected {
		t.Errorf("oauth_signature = %q, want %q", oauthParams["oauth_signature"], expected)
	}
}

func BenchmarkPercentEncode(b *testing.B) {
	b.ReportAllocs()
	s := "https://api.flickr.com/services/rest?method=flickr.photos.search&text=hello+world"
	for b.Loop() {
		PercentEncode(s)
	}
}

func BenchmarkNormalizeParams(b *testing.B) {
	b.ReportAllocs()
	params := map[string][]string{
		"oauth_consumer_key":     {"abcdef1234567890"},
		"oauth_nonce":            {"kllo9940pd9333jh"},
		"oauth_signature_method": {"HMAC-SHA1"},
		"oauth_timestamp":        {"1191242096"},
		"oauth_version":          {"1.0"},
	}
	for b.Loop() {
		NormalizeParams(params)
	}
}

func BenchmarkHMACSHA1Signature(b *testing.B) {
	b.ReportAllocs()
	baseString := "POST&https%3A%2F%2Fapi.flickr.com%2Fservices%2Frest&method%3Dflickr.photos.search%26oauth_consumer_key%3Dabcdef1234567890%26oauth_nonce%3Dkllo9940pd9333jh%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1191242096%26oauth_version%3D1.0"
	signingKey := "consumer_secret&token_secret"
	for b.Loop() {
		HMACSHA1Signature(baseString, signingKey)
	}
}
