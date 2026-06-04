package flickr

import "testing"

func TestFlickrError(t *testing.T) {
	err := &FlickrError{
		Method: "flickr.test.echo",
		Code:   1,
		Msg:    "Not found",
		Stat:   "fail",
	}

	expected := "flickr API error (method=flickr.test.echo, code=1): Not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestFlickrErrorIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		retryable bool
	}{
		{"rate limit 429", 429, true},
		{"server error 500", 500, true},
		{"bad gateway 502", 502, true},
		{"service unavailable 503", 503, true},
		{"not found 404", 404, false},
		{"bad request 400", 400, false},
		{"flickr API error code 1", 1, false},
		{"flickr API error code 96", 96, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &FlickrError{Method: "test", Code: tt.code, Msg: "x"}
			if got := err.IsRetryable(); got != tt.retryable {
				t.Errorf("IsRetryable() for code %d = %v, want %v", tt.code, got, tt.retryable)
			}
		})
	}
}
