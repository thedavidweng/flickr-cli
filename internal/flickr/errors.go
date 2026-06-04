package flickr

import (
	"fmt"
	"time"
)

// FlickrError represents an error returned by the Flickr API.
type FlickrError struct {
	Method     string
	Code       int
	Msg        string
	Stat       string
	RetryAfter time.Duration // populated for HTTP 429 responses
}

func (e *FlickrError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("flickr API error (method=%s, code=%d): %s [retry after %s]", e.Method, e.Code, e.Msg, e.RetryAfter)
	}
	return fmt.Sprintf("flickr API error (method=%s, code=%d): %s", e.Method, e.Code, e.Msg)
}

// IsRetryable reports whether the error is caused by a transient condition
// (HTTP 429 rate-limit or 5xx server error) that may succeed if retried.
func (e *FlickrError) IsRetryable() bool {
	return e.Code == 429 || e.Code >= 500
}
