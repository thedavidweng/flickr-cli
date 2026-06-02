package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

// Call records a single API call made to the fake server.
type Call struct {
	Method string
	Params map[string]string
}

// FakePhoto represents a photo in the fake server.
type FakePhoto struct {
	ID    string
	Title string
	Owner string
	Tags  string
}

// FakeAlbum represents an album in the fake server.
type FakeAlbum struct {
	ID          string
	Title       string
	Description string
	PhotoCount  int
	PrimaryID   string
}

// FakeFailure defines a failure for a specific method.
type FakeFailure struct {
	Code    int
	Message string
}

// FakeFlickr is a test double for the Flickr API.
type FakeFlickr struct {
	Server   *httptest.Server
	mu       sync.Mutex
	Calls    []Call
	Photos   map[string]FakePhoto
	Albums   map[string]FakeAlbum
	Failures map[string]FakeFailure
}

// NewFakeFlickr creates and starts a fake Flickr server.
func NewFakeFlickr(t *testing.T) *FakeFlickr {
	t.Helper()
	f := &FakeFlickr{
		Photos:   make(map[string]FakePhoto),
		Albums:   make(map[string]FakeAlbum),
		Failures: make(map[string]FakeFailure),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/services/rest/", f.handleREST)
	mux.HandleFunc("/services/upload/", f.handleUpload)
	mux.HandleFunc("/oauth/request_token", f.handleRequestToken)
	mux.HandleFunc("/oauth/access_token", f.handleAccessToken)
	mux.HandleFunc("/oauth/authorize", f.handleAuthorize)

	f.Server = httptest.NewServer(mux)
	return f
}

// Endpoints returns endpoints pointing to the fake server.
func (f *FakeFlickr) Endpoints() flickr.Endpoints {
	return flickr.Endpoints{
		REST:         f.Server.URL + "/services/rest/",
		Upload:       f.Server.URL + "/services/upload/",
		RequestToken: f.Server.URL + "/oauth/request_token",
		Authorize:    f.Server.URL + "/oauth/authorize",
		AccessToken:  f.Server.URL + "/oauth/access_token",
	}
}

// CountMethod returns the number of times a method was called.
func (f *FakeFlickr) CountMethod(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, c := range f.Calls {
		if c.Method == method {
			count++
		}
	}
	return count
}

// LastCall returns the last call to a method.
func (f *FakeFlickr) LastCall(method string) (Call, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.Calls) - 1; i >= 0; i-- {
		if f.Calls[i].Method == method {
			return f.Calls[i], true
		}
	}
	return Call{}, false
}

func (f *FakeFlickr) handleREST(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("method")
	if method == "" {
		method = r.URL.Query().Get("method")
	}

	f.mu.Lock()
	f.Calls = append(f.Calls, Call{
		Method: method,
		Params: extractParams(r),
	})

	// Check for configured failure
	if fail, ok := f.Failures[method]; ok {
		f.mu.Unlock()
		writeJSON(w, map[string]any{
			"stat":    "fail",
			"code":    fail.Code,
			"message": fail.Message,
		})
		return
	}
	f.mu.Unlock()

	switch method {
	case "flickr.test.login":
		writeJSON(w, map[string]any{
			"stat": "ok",
			"user": map[string]any{
				"id":       "test-user-123",
				"username": map[string]string{"_content": "testuser"},
			},
		})
	case "flickr.test.echo":
		writeJSON(w, map[string]any{"stat": "ok"})
	case "flickr.reflection.getMethods":
		writeJSON(w, map[string]any{
			"stat": "ok",
			"methods": map[string]any{
				"method": []map[string]string{
					{"_content": "flickr.test.login"},
					{"_content": "flickr.test.echo"},
				},
			},
		})
	case "flickr.photosets.getList":
		f.handleGetAlbums(w, r)
	case "flickr.photosets.getInfo":
		f.handleGetAlbumInfo(w, r)
	case "flickr.photosets.getPhotos":
		f.handleGetAlbumPhotos(w, r)
	case "flickr.photos.search":
		f.handlePhotoSearch(w, r)
	case "flickr.people.getPhotos":
		f.handleGetUserPhotos(w, r)
	case "flickr.photos.getInfo":
		f.handleGetPhotoInfo(w, r)
	case "flickr.photos.getSizes":
		f.handleGetPhotoSizes(w, r)
	case "flickr.photos.getAllContexts":
		writeJSON(w, map[string]any{"stat": "ok", "set": []any{}, "pool": []any{}})
	default:
		writeJSON(w, map[string]any{"stat": "ok"})
	}
}

func (f *FakeFlickr) handleUpload(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.Calls = append(f.Calls, Call{Method: "upload", Params: extractParams(r)})
	f.mu.Unlock()

	writeJSON(w, map[string]any{
		"stat":     "ok",
		"photoid":  "fake-photo-123",
		"ticketid": "ticket-1",
	})
}

func (f *FakeFlickr) handleRequestToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	fmt.Fprint(w, "oauth_token=req-token&oauth_token_secret=req-secret&oauth_callback_confirmed=true")
}

func (f *FakeFlickr) handleAccessToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	fmt.Fprint(w, "oauth_token=access-token&oauth_token_secret=access-secret&user_nsid=test-user-123&username=testuser&fullname=Test+User")
}

func (f *FakeFlickr) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	// In tests, we simulate immediate authorization
	verifier := "test-verifier"
	token := r.URL.Query().Get("oauth_token")
	redirectURI := fmt.Sprintf("%s?oauth_verifier=%s&oauth_token=%s", "http://localhost", verifier, token)
	http.Redirect(w, r, redirectURI, http.StatusFound)
}

func (f *FakeFlickr) handleGetAlbums(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	albums := make([]map[string]any, 0, len(f.Albums))
	for _, a := range f.Albums {
		albums = append(albums, map[string]any{
			"id":               a.ID,
			"title":            map[string]string{"_content": a.Title},
			"description":      map[string]string{"_content": a.Description},
			"photos":           a.PhotoCount,
			"primary_photo_id": a.PrimaryID,
		})
	}
	f.mu.Unlock()

	writeJSON(w, map[string]any{
		"stat": "ok",
		"photosets": map[string]any{
			"photoset": albums,
			"page":     1,
			"pages":    1,
			"perpage":  100,
			"total":    len(albums),
		},
	})
}

func (f *FakeFlickr) handleGetAlbumInfo(w http.ResponseWriter, r *http.Request) {
	photosetID := r.FormValue("photoset_id")
	if photosetID == "" {
		photosetID = r.URL.Query().Get("photoset_id")
	}

	f.mu.Lock()
	album, ok := f.Albums[photosetID]
	f.mu.Unlock()

	if !ok {
		writeJSON(w, map[string]any{"stat": "fail", "code": 1, "message": "Album not found"})
		return
	}

	writeJSON(w, map[string]any{
		"stat": "ok",
		"photoset": map[string]any{
			"id":          album.ID,
			"title":       map[string]string{"_content": album.Title},
			"description": map[string]string{"_content": album.Description},
			"photos":      album.PhotoCount,
		},
	})
}

func (f *FakeFlickr) handleGetAlbumPhotos(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"photoset": map[string]any{
			"photo":   []any{},
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   0,
		},
	})
}

func (f *FakeFlickr) handlePhotoSearch(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	photos := make([]map[string]any, 0, len(f.Photos))
	for _, p := range f.Photos {
		photos = append(photos, map[string]any{
			"id":    p.ID,
			"title": p.Title,
			"owner": p.Owner,
			"tags":  p.Tags,
		})
	}
	f.mu.Unlock()

	writeJSON(w, map[string]any{
		"stat": "ok",
		"photos": map[string]any{
			"photo":   photos,
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   len(photos),
		},
	})
}

func (f *FakeFlickr) handleGetUserPhotos(w http.ResponseWriter, r *http.Request) {
	f.handlePhotoSearch(w, r)
}

func (f *FakeFlickr) handleGetPhotoInfo(w http.ResponseWriter, r *http.Request) {
	photoID := r.FormValue("photo_id")
	if photoID == "" {
		photoID = r.URL.Query().Get("photo_id")
	}

	f.mu.Lock()
	photo, ok := f.Photos[photoID]
	f.mu.Unlock()

	if !ok {
		writeJSON(w, map[string]any{"stat": "fail", "code": 1, "message": "Photo not found"})
		return
	}

	writeJSON(w, map[string]any{
		"stat": "ok",
		"photo": map[string]any{
			"id":    photo.ID,
			"title": map[string]string{"_content": photo.Title},
			"owner": map[string]string{"nsid": photo.Owner},
			"tags":  map[string]any{"tag": []any{}},
		},
	})
}

func (f *FakeFlickr) handleGetPhotoSizes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"sizes": map[string]any{
			"size": []map[string]any{
				{"label": "Original", "width": 4000, "height": 3000, "source": "https://example.com/photo.jpg", "url": "https://example.com/photo.jpg", "media": "photo"},
				{"label": "Large", "width": 1024, "height": 768, "source": "https://example.com/photo_l.jpg", "url": "https://example.com/photo_l.jpg", "media": "photo"},
			},
		},
	})
}

func extractParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for k, vs := range r.Form {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	for k, vs := range r.URL.Query() {
		if _, exists := params[k]; !exists && len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	return params
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// Client creates a flickr.Client pointing at the fake server.
func (f *FakeFlickr) Client() *flickr.Client {
	return &flickr.Client{
		APIKey:      "test-api-key",
		APISecret:   "test-api-secret",
		OAuthToken:  "test-token",
		OAuthSecret: "test-secret",
		HTTP:        f.Server.Client(),
		UserAgent:   "flickr-cli-test",
		Endpoints:   f.Endpoints(),
	}
}
