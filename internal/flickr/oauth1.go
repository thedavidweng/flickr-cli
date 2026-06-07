package flickr

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// OAuthCredentials holds OAuth 1.0a consumer and token credentials.
type OAuthCredentials struct {
	ConsumerKey    string
	ConsumerSecret string
	Token          string
	TokenSecret    string
}

// OAuthSigner signs requests using OAuth 1.0a HMAC-SHA1.
type OAuthSigner struct {
	Creds OAuthCredentials
	Now   func() time.Time
	Nonce func() string
}

// NewOAuthSigner creates a signer with default time and nonce functions.
func NewOAuthSigner(creds OAuthCredentials) OAuthSigner {
	return OAuthSigner{
		Creds: creds,
		Now:   time.Now,
		Nonce: func() string {
			return fmt.Sprintf("%016x", rand.Int63())
		},
	}
}

// PercentEncode encodes a string per RFC3986.
// A-Z a-z 0-9 - . _ ~ remain unescaped, space encodes as %20.
func PercentEncode(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~' {
			buf.WriteByte(c)
		} else {
			fmt.Fprintf(&buf, "%%%02X", c)
		}
	}
	return buf.String()
}

// NormalizeParams normalizes parameters per OAuth spec.
// Parameters are sorted by encoded key, then by encoded value.
func NormalizeParams(params map[string][]string) string {
	type kv struct{ k, v string }
	var pairs []kv
	for k, vs := range params {
		ek := PercentEncode(k)
		for _, v := range vs {
			pairs = append(pairs, kv{ek, PercentEncode(v)})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].k != pairs[j].k {
			return pairs[i].k < pairs[j].k
		}
		return pairs[i].v < pairs[j].v
	})
	var buf strings.Builder
	for i, p := range pairs {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(p.k)
		buf.WriteByte('=')
		buf.WriteString(p.v)
	}
	return buf.String()
}

// SignatureBaseString constructs the signature base string.
func SignatureBaseString(method, baseURL string, params map[string][]string) string {
	return strings.ToUpper(method) + "&" + PercentEncode(baseURL) + "&" + PercentEncode(NormalizeParams(params))
}

// SigningKey constructs the signing key.
func SigningKey(consumerSecret, tokenSecret string) string {
	return PercentEncode(consumerSecret) + "&" + PercentEncode(tokenSecret)
}

// HMACSHA1Signature computes the HMAC-SHA1 signature.
func HMACSHA1Signature(baseString, signingKey string) string {
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// OAuthParams returns the standard OAuth parameters.
func (s OAuthSigner) OAuthParams() map[string]string {
	return map[string]string{
		"oauth_consumer_key":     s.Creds.ConsumerKey,
		"oauth_nonce":            s.Nonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%d", s.Now().Unix()),
		"oauth_version":          "1.0",
	}
}

// Sign signs a request and returns the OAuth parameters including signature.
// The returned map includes all OAuth protocol parameters (consumer_key, nonce,
// timestamp, signature, etc.) plus any params that start with "oauth_" from the
// input params (e.g. oauth_callback, oauth_verifier).
func (s OAuthSigner) Sign(method, baseURL string, params map[string][]string) (map[string]string, error) {
	oauthParams := s.OAuthParams()

	// Add token if present
	if s.Creds.Token != "" {
		oauthParams["oauth_token"] = s.Creds.Token
	}

	// Promote oauth_* params from the input into both the signature base
	// string and the returned OAuth params (they belong in the Authorization
	// header, not the request body).
	for k, vs := range params {
		if len(vs) > 0 && len(k) > 6 && k[:6] == "oauth_" {
			oauthParams[k] = vs[0]
		}
	}

	// Merge OAuth params into signature params
	sigParams := make(map[string][]string)
	for k, vs := range params {
		sigParams[k] = vs
	}
	for k, v := range oauthParams {
		if k == "oauth_signature" {
			continue
		}
		sigParams[k] = []string{v}
	}

	baseString := SignatureBaseString(method, baseURL, sigParams)
	signingKey := SigningKey(s.Creds.ConsumerSecret, s.Creds.TokenSecret)
	oauthParams["oauth_signature"] = HMACSHA1Signature(baseString, signingKey)

	return oauthParams, nil
}

// AuthorizationHeader formats OAuth parameters as an Authorization header value.
func AuthorizationHeader(oauthParams map[string]string) string {
	keys := make([]string, 0, len(oauthParams))
	for k := range oauthParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, PercentEncode(k)+"=\""+PercentEncode(oauthParams[k])+"\"")
	}
	return "OAuth " + strings.Join(parts, ", ")
}
