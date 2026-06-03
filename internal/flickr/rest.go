package flickr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// flickrResponse is the common wrapper for Flickr REST responses.
type flickrResponse struct {
	Stat    string          `json:"stat"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Raw     json.RawMessage `json:"-"`
}

// CallRaw makes a REST API call and returns the raw JSON response.
// It retries on HTTP 5xx status codes up to c.Retries times with exponential backoff.
func (c *Client) CallRaw(ctx context.Context, method string, params map[string]string) (json.RawMessage, error) {
	maxAttempts := 1 + c.Retries

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		raw, err, retry := c.callRawOnce(ctx, method, params)
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, lastErr
}

// callRawOnce performs a single REST API call attempt.
// It returns (raw, nil, false) on success, or (nil, err, true) if the error is retryable.
func (c *Client) callRawOnce(ctx context.Context, method string, params map[string]string) (json.RawMessage, error, bool) {
	// Build form values
	form := url.Values{}
	form.Set("method", method)
	form.Set("api_key", c.APIKey)
	form.Set("format", "json")
	form.Set("nojsoncallback", "1")
	for k, v := range params {
		form.Set(k, v)
	}

	var (
		req *http.Request
		err error
	)

	if c.IsAuthenticated() {
		// Authenticated call: POST with OAuth signature
		sigParams := make(map[string][]string)
		for k, vs := range form {
			sigParams[k] = vs
		}

		signer := c.Signer()
		oauthParams, signErr := signer.Sign("POST", c.Endpoints.REST, sigParams)
		if signErr != nil {
			return nil, fmt.Errorf("signing request: %w", signErr), false
		}

		req, err = http.NewRequestWithContext(ctx, "POST", c.Endpoints.REST, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err), false
		}
		req.Header.Set("Authorization", AuthorizationHeader(oauthParams))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		// Unauthenticated call: GET with query params
		reqURL := c.Endpoints.REST + "?" + form.Encode()
		req, err = http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err), false
		}
	}

	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call %s: %w", method, err), false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err), false
	}

	if resp.StatusCode != http.StatusOK {
		flickrErr := &FlickrError{
			Method: method,
			Code:   resp.StatusCode,
			Msg:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			Stat:   "http_error",
		}
		retryable := resp.StatusCode >= 500
		return nil, flickrErr, retryable
	}

	// Check for Flickr API error
	var fr flickrResponse
	if err := json.Unmarshal(body, &fr); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err), false
	}
	if fr.Stat == "fail" {
		return nil, &FlickrError{
			Method: method,
			Code:   fr.Code,
			Msg:    fr.Message,
			Stat:   fr.Stat,
		}, false
	}

	return body, nil, false
}

// Call makes a REST API call and decodes the response into out.
func (c *Client) Call(ctx context.Context, method string, params map[string]string, out any) error {
	raw, err := c.CallRaw(ctx, method, params)
	if err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decoding response for %s: %w", method, err)
		}
	}
	return nil
}

// TestLogin tests if the current credentials are valid.
func (c *Client) TestLogin(ctx context.Context) (*LoginInfo, error) {
	var result struct {
		User struct {
			ID       string `json:"id"`
			Username struct {
				Content string `json:"_content"`
			} `json:"username"`
		} `json:"user"`
	}
	if err := c.Call(ctx, "flickr.test.login", nil, &result); err != nil {
		return nil, err
	}
	return &LoginInfo{
		UserNSID: result.User.ID,
		Username: result.User.Username.Content,
	}, nil
}

// TestEcho tests the API connection.
func (c *Client) TestEcho(ctx context.Context) error {
	return c.Call(ctx, "flickr.test.echo", nil, nil)
}
