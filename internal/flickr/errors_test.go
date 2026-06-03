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
