package flickr

import (
	"context"
	"encoding/json"
)

// FlickrAPI is the seam for all Flickr API access.
// *Client satisfies this interface implicitly.
// Packages that only call the API (backup, upload, cache, piwigo)
// depend on this interface rather than the concrete *Client,
// which enables in-memory test doubles without an HTTP server.
type FlickrAPI interface {
	// REST
	Call(ctx context.Context, method string, params map[string]string, out any) error
	CallRaw(ctx context.Context, method string, params map[string]string) (json.RawMessage, error)
	TestLogin(ctx context.Context) (*LoginInfo, error)
	TestEcho(ctx context.Context) error

	// Photo operations
	GetSizes(ctx context.Context, photoID string) ([]Size, error)
	GetVideoStreams(ctx context.Context, photoID string) ([]VideoStream, error)
	GetExif(ctx context.Context, photoID string) (*ExifData, error)
	Upload(ctx context.Context, filePath string, opts *UploadOptions) (*UploadResult, error)
	AddToAlbum(ctx context.Context, albumID, photoID string) error

	// Auth
	IsAuthenticated() bool
	RequestToken(ctx context.Context, callback string) (*RequestTokenResponse, error)
	AuthorizationURL(token, perms string) string
	AccessToken(ctx context.Context, requestToken, requestTokenSecret, verifier string) (*AccessTokenResponse, error)

	// Reflection
	GetMethods(ctx context.Context) (json.RawMessage, error)
	GetMethodInfo(ctx context.Context, methodName string) (json.RawMessage, error)
}

// Verify that *Client satisfies FlickrAPI at compile time.
var _ FlickrAPI = (*Client)(nil)
