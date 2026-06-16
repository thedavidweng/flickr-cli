package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// FakeSize represents a photo size in the fake server.
type FakeSize struct {
	Label  string
	Source string
	Width  int
	Height int
}

// FakeFlickr is a test double for the Flickr API.
type FakeFlickr struct {
	Server      *httptest.Server
	mu          sync.Mutex
	Calls       []Call
	Photos      map[string]FakePhoto
	Albums      map[string]FakeAlbum
	AlbumPhotos map[string][]string // album ID -> ordered photo IDs
	PhotoSizes  map[string][]FakeSize // photo ID -> custom sizes (nil = use defaults)
	Failures    map[string]FakeFailure
}

// NewFakeFlickr creates and starts a fake Flickr server.
func NewFakeFlickr(t *testing.T) *FakeFlickr {
	t.Helper()
	f := &FakeFlickr{
		Photos:      make(map[string]FakePhoto),
		Albums:      make(map[string]FakeAlbum),
		AlbumPhotos: make(map[string][]string),
		PhotoSizes:  make(map[string][]FakeSize),
		Failures:    make(map[string]FakeFailure),
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
	case "flickr.photos.getExif":
		f.handleGetExif(w, r)
	case "flickr.video.getStreamInfo":
		f.handleGetVideoStreams(w, r)
	case "flickr.photos.getAllContexts":
		writeJSON(w, map[string]any{"stat": "ok", "set": []any{}, "pool": []any{}})
	case "flickr.favorites.getList":
		f.handleGetFavorites(w, r)
	case "flickr.galleries.getList":
		f.handleGetGalleries(w, r)
	case "flickr.galleries.getPhotos":
		f.handleGetGalleryPhotos(w, r)
	case "flickr.groups.getList":
		f.handleGetGroups(w, r)
	case "flickr.groups.search":
		f.handleSearchGroups(w, r)
	case "flickr.contacts.getList":
		f.handleGetContacts(w, r)
	case "flickr.stats.getPopularPhotos":
		f.handleGetPopularPhotos(w, r)
	case "flickr.urls.lookupUser":
		f.handleLookupUser(w, r)
	case "flickr.photos.comments.getList":
		f.handleGetComments(w, r)
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
	_, _ = fmt.Fprint(w, "oauth_token=req-token&oauth_token_secret=req-secret&oauth_callback_confirmed=true")
}

func (f *FakeFlickr) handleAccessToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	_, _ = fmt.Fprint(w, "oauth_token=access-token&oauth_token_secret=access-secret&user_nsid=test-user-123&username=testuser&fullname=Test+User")
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
	photosetID := r.FormValue("photoset_id")
	if photosetID == "" {
		photosetID = r.URL.Query().Get("photoset_id")
	}

	f.mu.Lock()
	photoIDs := f.AlbumPhotos[photosetID]
	photos := make([]map[string]any, 0, len(photoIDs))
	for _, pid := range photoIDs {
		if p, ok := f.Photos[pid]; ok {
			photos = append(photos, map[string]any{
				"id":    p.ID,
				"title": p.Title,
				"owner": p.Owner,
			})
		} else {
			photos = append(photos, map[string]any{
				"id":    pid,
				"title": "",
			})
		}
	}
	f.mu.Unlock()

	writeJSON(w, map[string]any{
		"stat": "ok",
		"photoset": map[string]any{
			"id":      photosetID,
			"photo":   photos,
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   len(photos),
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

	// Parse tags from the FakePhoto into the expected format.
	var tags []map[string]any
	if photo.Tags != "" {
		for _, raw := range strings.Split(photo.Tags, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			tags = append(tags, map[string]any{
				"raw":     raw,
				"machine": 0,
			})
		}
	}

	writeJSON(w, map[string]any{
		"stat": "ok",
		"photo": map[string]any{
			"id":    photo.ID,
			"title": map[string]string{"_content": photo.Title},
			"owner": map[string]string{"nsid": photo.Owner},
			"tags":  map[string]any{"tag": tags},
		},
	})
}

func (f *FakeFlickr) handleGetPhotoSizes(w http.ResponseWriter, r *http.Request) {
	photoID := r.FormValue("photo_id")
	if photoID == "" {
		photoID = r.URL.Query().Get("photo_id")
	}

	f.mu.Lock()
	customSizes, hasCustom := f.PhotoSizes[photoID]
	f.mu.Unlock()

	var sizes []map[string]any
	if hasCustom && len(customSizes) > 0 {
		for _, s := range customSizes {
			sizes = append(sizes, map[string]any{
				"label":  s.Label,
				"width":  s.Width,
				"height": s.Height,
				"source": s.Source,
				"url":    s.Source,
				"media":  "photo",
			})
		}
	} else {
		sizes = []map[string]any{
			{"label": "Original", "width": 4000, "height": 3000, "source": "https://example.com/photo.jpg", "url": "https://example.com/photo.jpg", "media": "photo"},
			{"label": "Large", "width": 1024, "height": 768, "source": "https://example.com/photo_l.jpg", "url": "https://example.com/photo_l.jpg", "media": "photo"},
		}
	}

	writeJSON(w, map[string]any{
		"stat":  "ok",
		"sizes": map[string]any{"size": sizes},
	})
}

func (f *FakeFlickr) handleGetExif(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"photo": map[string]any{
			"id": r.FormValue("photo_id"),
			"tag": []map[string]any{
				{"tagspace": "EXIF", "tagspaceid": 0, "tag": "Make", "label": "Make", "raw": "Canon", "_content": "Canon"},
				{"tagspace": "EXIF", "tagspaceid": 0, "tag": "Model", "label": "Model", "raw": "EOS R5", "_content": "EOS R5"},
			},
		},
	})
}

func (f *FakeFlickr) handleGetVideoStreams(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"streams": map[string]any{
			"stream": []map[string]any{
				{"type": "1080p", "width": 1920, "height": 1080, "source": "https://example.com/video_1080.mp4"},
				{"type": "720p", "width": 1280, "height": 720, "source": "https://example.com/video_720.mp4"},
			},
		},
	})
}

func (f *FakeFlickr) handleGetFavorites(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	photos := make([]map[string]any, 0, len(f.Photos))
	for _, p := range f.Photos {
		photos = append(photos, map[string]any{
			"id":    p.ID,
			"title": p.Title,
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

func (f *FakeFlickr) handleGetGalleries(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"galleries": map[string]any{
			"gallery": []map[string]any{
				{"id": "gallery-1", "title": "Test Gallery", "description": "A gallery", "count_photos": 5},
			},
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   1,
		},
	})
}

func (f *FakeFlickr) handleGetGalleryPhotos(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"gallery": map[string]any{
			"id":    "gallery-1",
			"title": "Test Gallery",
		},
		"photos": map[string]any{
			"photo":   []any{},
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   0,
		},
	})
}

func (f *FakeFlickr) handleGetGroups(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"groups": map[string]any{
			"group": []map[string]any{
				{"nsid": "group-1", "name": "Test Group", "members": 100},
			},
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   1,
		},
	})
}

func (f *FakeFlickr) handleSearchGroups(w http.ResponseWriter, r *http.Request) {
	f.handleGetGroups(w, r)
}

func (f *FakeFlickr) handleGetContacts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"contacts": map[string]any{
			"contact": []map[string]any{
				{"nsid": "user-1", "username": "testuser", "realname": "Test User"},
			},
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   1,
		},
	})
}

func (f *FakeFlickr) handleGetPopularPhotos(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"photos": map[string]any{
			"photo": []map[string]any{
				{"id": "p1", "title": "Popular Photo", "stats": map[string]any{"views": 1000}},
			},
			"page":    1,
			"pages":   1,
			"perpage": 100,
			"total":   1,
		},
	})
}

func (f *FakeFlickr) handleLookupUser(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"user": map[string]any{
			"id":       "user-123",
			"username": "testuser",
		},
	})
}

func (f *FakeFlickr) handleGetComments(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"stat": "ok",
		"comments": map[string]any{
			"comment": []map[string]any{
				{"id": "comment-1", "authorname": "testuser", "_content": "Great photo!", "datecreate": "2024-01-01"},
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
	_ = json.NewEncoder(w).Encode(v)
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
