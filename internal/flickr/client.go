package flickr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client is the Flickr API client.
type Client struct {
	APIKey          string
	APISecret       string
	OAuthToken      string
	OAuthSecret     string
	HTTP            *http.Client
	UserAgent       string
	Endpoints       Endpoints
	Retries         int
	RequestInterval time.Duration // minimum interval between API calls (0 = no delay)
	lastCallTime    time.Time
	rateMu          sync.Mutex
}

// NewClient creates a new Flickr client with default settings.
func NewClient(apiKey, apiSecret, oauthToken, oauthSecret string) *Client {
	return &Client{
		APIKey:      apiKey,
		APISecret:   apiSecret,
		OAuthToken:  oauthToken,
		OAuthSecret: oauthSecret,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: "flickr-cli/1.0",
		Endpoints: DefaultEndpoints(),
	}
}

// IsAuthenticated returns true if the client has OAuth credentials.
func (c *Client) IsAuthenticated() bool {
	return c.OAuthToken != "" && c.OAuthSecret != ""
}

// waitForRateLimit enforces the minimum interval between API calls.
// It respects context cancellation to avoid hanging on Ctrl+C.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	if c.RequestInterval <= 0 {
		return nil
	}
	c.rateMu.Lock()
	defer c.rateMu.Unlock()
	elapsed := time.Since(c.lastCallTime)
	if elapsed < c.RequestInterval {
		wait := c.RequestInterval - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	c.lastCallTime = time.Now()
	return nil
}

// Signer returns an OAuthSigner for this client.
func (c *Client) Signer() OAuthSigner {
	return NewOAuthSigner(OAuthCredentials{
		ConsumerKey:    c.APIKey,
		ConsumerSecret: c.APISecret,
		Token:          c.OAuthToken,
		TokenSecret:    c.OAuthSecret,
	})
}

// RequestToken obtains a request token from Flickr.
func (c *Client) RequestToken(ctx context.Context, callback string) (*RequestTokenResponse, error) {
	params := map[string][]string{
		"oauth_callback": {callback},
	}

	signer := c.Signer()
	oauthParams, err := signer.Sign("POST", c.Endpoints.RequestToken, params)
	if err != nil {
		return nil, fmt.Errorf("signing request token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoints.RequestToken, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", AuthorizationHeader(oauthParams))
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request token failed: %s %s", resp.Status, string(body))
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &RequestTokenResponse{
		Token:             values.Get("oauth_token"),
		TokenSecret:       values.Get("oauth_token_secret"),
		CallbackConfirmed: values.Get("oauth_callback_confirmed") == "true",
	}, nil
}

// AuthorizationURL returns the URL for user authorization.
func (c *Client) AuthorizationURL(token, perms string) string {
	params := url.Values{
		"oauth_token": {token},
	}
	if perms != "" {
		params.Set("perms", perms)
	}
	return c.Endpoints.Authorize + "?" + params.Encode()
}

// AccessToken exchanges a request token and verifier for an access token.
func (c *Client) AccessToken(ctx context.Context, requestToken, requestTokenSecret, verifier string) (*AccessTokenResponse, error) {
	// Create a temporary client with request token credentials
	tempClient := &Client{
		APIKey:      c.APIKey,
		APISecret:   c.APISecret,
		OAuthToken:  requestToken,
		OAuthSecret: requestTokenSecret,
		HTTP:        c.HTTP,
		UserAgent:   c.UserAgent,
		Endpoints:   c.Endpoints,
	}

	params := map[string][]string{
		"oauth_verifier": {verifier},
	}

	signer := tempClient.Signer()
	oauthParams, err := signer.Sign("POST", c.Endpoints.AccessToken, params)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	// Add verifier to form body
	form := url.Values{
		"oauth_verifier": {verifier},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoints.AccessToken, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", AuthorizationHeader(oauthParams))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("access token failed: %s %s", resp.Status, string(body))
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &AccessTokenResponse{
		Token:       values.Get("oauth_token"),
		TokenSecret: values.Get("oauth_token_secret"),
		UserNSID:    values.Get("user_nsid"),
		Username:    values.Get("username"),
		FullName:    values.Get("fullname"),
	}, nil
}
