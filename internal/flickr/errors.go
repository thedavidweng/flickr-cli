package flickr

import "fmt"

// FlickrError represents an error returned by the Flickr API.
type FlickrError struct {
	Method string
	Code   int
	Msg    string
	Stat   string
}

func (e *FlickrError) Error() string {
	return fmt.Sprintf("flickr API error (method=%s, code=%d): %s", e.Method, e.Code, e.Msg)
}
