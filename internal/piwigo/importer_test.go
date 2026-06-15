package piwigo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// ---------------------------------------------------------------------------
// Minimal FlickrAPI mock
// ---------------------------------------------------------------------------

type mockFlickr struct {
	uploaded     []string
	albums       map[string][]string
	failUpload   bool
	onCall       func(method string, params map[string]string) error
	uploadCalled atomic.Int32
}

func newMockFlickr() *mockFlickr {
	return &mockFlickr{
		uploaded: []string{
			"flickr_photo_1", "flickr_photo_2", "flickr_photo_3",
			"flickr_photo_4", "flickr_photo_5",
		},
		albums: make(map[string][]string),
	}
}

func (m *mockFlickr) Upload(_ context.Context, _ string, _ *flickr.UploadOptions) (*flickr.UploadResult, error) {
	m.uploadCalled.Add(1)
	if m.failUpload {
		return nil, fmt.Errorf("upload error")
	}
	if len(m.uploaded) == 0 {
		return nil, fmt.Errorf("no more photo IDs")
	}
	id := m.uploaded[0]
	m.uploaded = m.uploaded[1:]
	return &flickr.UploadResult{PhotoID: id}, nil
}

func (m *mockFlickr) AddToAlbum(_ context.Context, albumID, photoID string) error {
	m.albums[albumID] = append(m.albums[albumID], photoID)
	return nil
}

func (m *mockFlickr) Call(_ context.Context, method string, params map[string]string, _ any) error {
	if m.onCall != nil {
		return m.onCall(method, params)
	}
	return nil
}
func (m *mockFlickr) CallRaw(_ context.Context, _ string, _ map[string]string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockFlickr) TestLogin(_ context.Context) (*flickr.LoginInfo, error) {
	return &flickr.LoginInfo{}, nil
}
func (m *mockFlickr) TestEcho(_ context.Context) error                            { return nil }
func (m *mockFlickr) GetSizes(_ context.Context, _ string) ([]flickr.Size, error) { return nil, nil }
func (m *mockFlickr) GetVideoStreams(_ context.Context, _ string) ([]flickr.VideoStream, error) {
	return nil, nil
}
func (m *mockFlickr) GetExif(_ context.Context, _ string) (*flickr.ExifData, error) { return nil, nil }
func (m *mockFlickr) IsAuthenticated() bool                                         { return true }
func (m *mockFlickr) RequestToken(_ context.Context, _ string) (*flickr.RequestTokenResponse, error) {
	return nil, nil
}
func (m *mockFlickr) AuthorizationURL(_, _ string) string { return "" }
func (m *mockFlickr) AccessToken(_ context.Context, _, _, _ string) (*flickr.AccessTokenResponse, error) {
	return nil, nil
}
func (m *mockFlickr) GetMethods(_ context.Context) (json.RawMessage, error) { return nil, nil }
func (m *mockFlickr) GetMethodInfo(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, nil
}

var _ flickr.FlickrAPI = (*mockFlickr)(nil)

// ---------------------------------------------------------------------------
// Combined Piwigo + image-download fake server
// ---------------------------------------------------------------------------

type testServer struct {
	categories  []Category
	images      map[string][]ImageInfo
	existingMD5 map[string]bool
	loginOK     bool
	downloadOK  bool
}

func (ts *testServer) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/ws.php":
			ts.handleAPI(w, r)
		case strings.HasPrefix(r.URL.Path, "/upload/"):
			ts.handleDownload(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

func (ts *testServer) handleAPI(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	method := r.FormValue("method")
	w.Header().Set("Content-Type", "application/json")

	switch method {
	case "pwg.session.login":
		if !ts.loginOK {
			_, _ = fmt.Fprint(w, `{"stat":"fail","message":"bad credentials"}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"stat":"ok","result":{"token":"tok123"}}`)

	case "pwg.categories.getList":
		data, _ := json.Marshal(ts.categories)
		_, _ = fmt.Fprintf(w, `{"stat":"ok","result":%s}`, data)

	case "pwg.categories.getImages":
		catID := r.FormValue("category_id")
		imgs := ts.images[catID]
		data, _ := json.Marshal(imgs)
		_, _ = fmt.Fprintf(w, `{"stat":"ok","result":%s,"paging":{"total_pages":1}}`, data)

	case "pwg.images.exist":
		md5List := r.FormValue("md5sum_list")
		results := make(map[string]bool)
		for _, md5 := range strings.Split(md5List, ",") {
			md5 = strings.TrimSpace(md5)
			if md5 != "" {
				results[md5] = ts.existingMD5[md5]
			}
		}
		data, _ := json.Marshal(results)
		_, _ = fmt.Fprintf(w, `{"stat":"ok","result":%s}`, data)

	default:
		_, _ = fmt.Fprint(w, `{"stat":"fail","message":"unknown method"}`)
	}
}

func (ts *testServer) handleDownload(w http.ResponseWriter, _ *http.Request) {
	if !ts.downloadOK {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	_, _ = w.Write([]byte("fakeimage"))
}

func silentEvents() *output.EventWriter {
	return &output.EventWriter{Enabled: false}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestImportSuccess(t *testing.T) {
	ts := &testServer{
		loginOK:    true,
		downloadOK: true,
		categories: []Category{
			{ID: "cat1", Name: "Vacation", NbImages: 2},
			{ID: "cat2", Name: "Nature", NbImages: 1},
		},
		images: map[string][]ImageInfo{
			"cat1": {
				{ID: "img1", File: "photo1.jpg", Name: "Beach", Comment: "sunny", MD5Sum: "aaa"},
				{ID: "img2", File: "photo2.jpg", Name: "Mountains", Comment: "snow", MD5Sum: "bbb"},
			},
			"cat2": {
				{ID: "img3", File: "photo3.jpg", Name: "Flower", Comment: "red", MD5Sum: "ccc"},
			},
		},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{
		URL:      srv.URL,
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Planned != 3 {
		t.Errorf("Planned = %d, want 3", summary.Planned)
	}
	if summary.Succeeded != 3 {
		t.Errorf("Succeeded = %d, want 3", summary.Succeeded)
	}
	if summary.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", summary.Skipped)
	}
	if summary.Failed != 0 {
		t.Errorf("Failed = %d, want 0", summary.Failed)
	}
	if mock.uploadCalled.Load() != 3 {
		t.Errorf("Upload called %d times, want 3", mock.uploadCalled.Load())
	}
}

func TestImportDedupeSkip(t *testing.T) {
	ts := &testServer{
		loginOK:    true,
		downloadOK: true,
		categories: []Category{
			{ID: "cat1", Name: "Album1", NbImages: 2},
		},
		images: map[string][]ImageInfo{
			"cat1": {
				{ID: "img1", File: "dup.jpg", Name: "Dup", MD5Sum: "already_here"},
				{ID: "img2", File: "new.jpg", Name: "New", MD5Sum: "not_here"},
			},
		},
		existingMD5: map[string]bool{"already_here": true},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{
		URL:    srv.URL,
		Dedupe: "checksum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Planned != 2 {
		t.Errorf("Planned = %d, want 2", summary.Planned)
	}
	if summary.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", summary.Skipped)
	}
	if summary.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", summary.Succeeded)
	}
}

func TestImportLimit(t *testing.T) {
	ts := &testServer{
		loginOK:    true,
		downloadOK: true,
		categories: []Category{
			{ID: "cat1", Name: "Big", NbImages: 5},
		},
		images: map[string][]ImageInfo{
			"cat1": {
				{ID: "img1", File: "a.jpg", Name: "A", MD5Sum: "a1"},
				{ID: "img2", File: "b.jpg", Name: "B", MD5Sum: "b1"},
				{ID: "img3", File: "c.jpg", Name: "C", MD5Sum: "c1"},
				{ID: "img4", File: "d.jpg", Name: "D", MD5Sum: "d1"},
				{ID: "img5", File: "e.jpg", Name: "E", MD5Sum: "e1"},
			},
		},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{
		URL:   srv.URL,
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Succeeded != 2 {
		t.Errorf("Succeeded = %d, want 2", summary.Succeeded)
	}
	// Planned counts every image iterated before the limit is hit:
	// img1 -> succeeds (count=1), img2 -> succeeds (count=2), then limit stops.
	if summary.Planned < 2 {
		t.Errorf("Planned = %d, want at least 2", summary.Planned)
	}
}

func TestDownloadToTemp(t *testing.T) {
	body := []byte("image-data-payload-12345")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	path, err := downloadToTemp(context.Background(), srv.URL+"/upload/photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = os.Remove(path) }()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading temp file: %v", err)
	}
	if string(data) != string(body) {
		t.Errorf("file content = %q, want %q", data, body)
	}
}

func TestDownloadToTempHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := downloadToTemp(context.Background(), srv.URL+"/missing.jpg")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention 404, got: %v", err)
	}
}

func TestImportLoginError(t *testing.T) {
	ts := &testServer{loginOK: false}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	_, err := imp.Import(context.Background(), &ImportOptions{
		URL:      srv.URL,
		Username: "bad",
		Password: "creds",
	})
	if err == nil {
		t.Fatal("expected error for login failure, got nil")
	}
	if !strings.Contains(err.Error(), "login") {
		t.Errorf("error should mention login, got: %v", err)
	}
}

func TestImportSkipEmptyCategory(t *testing.T) {
	ts := &testServer{
		loginOK: true,
		categories: []Category{
			{ID: "empty", Name: "Empty", NbImages: 0},
		},
		images: map[string][]ImageInfo{},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{URL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Planned != 0 {
		t.Errorf("Planned = %d, want 0", summary.Planned)
	}
	if mock.uploadCalled.Load() != 0 {
		t.Errorf("Upload called %d times, want 0", mock.uploadCalled.Load())
	}
}

func TestImportWithAlbums(t *testing.T) {
	ts := &testServer{
		loginOK:    true,
		downloadOK: true,
		categories: []Category{
			{ID: "c1", Name: "Travel", NbImages: 1},
		},
		images: map[string][]ImageInfo{
			"c1": {
				{
					ID:     "img1",
					File:   "trip.jpg",
					Name:   "Trip",
					MD5Sum: "abc",
					Categories: []struct {
						ID string `json:"id"`
					}{{ID: "c1"}},
					Tags: []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					}{{ID: "t1", Name: "vacation"}},
				},
			},
		},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{
		URL:         srv.URL,
		AlbumPrefix: "P/",
		ImportAlbum: "AllImports",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", summary.Succeeded)
	}
	if _, ok := mock.albums["AllImports"]; !ok {
		t.Error("expected AllImports album to be created")
	}
	if _, ok := mock.albums["P/Travel"]; !ok {
		t.Error("expected P/Travel album to be created")
	}
}

func TestImportWithGeo(t *testing.T) {
	ts := &testServer{
		loginOK:    true,
		downloadOK: true,
		categories: []Category{
			{ID: "c1", Name: "Geo", NbImages: 1},
		},
		images: map[string][]ImageInfo{
			"c1": {
				{ID: "img1", File: "geo.jpg", Name: "Geo", MD5Sum: "geo1", Latitude: 40.7128, Longitude: -74.0060},
			},
		},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	var geoCalled bool
	mock := newMockFlickr()
	mock.onCall = func(method string, params map[string]string) error {
		if method == "flickr.photos.geo.setLocation" {
			geoCalled = true
			if params["lat"] == "" || params["lon"] == "" {
				t.Error("expected lat/lon in geo params")
			}
		}
		return nil
	}

	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{URL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", summary.Succeeded)
	}
	if !geoCalled {
		t.Error("expected flickr.photos.geo.setLocation to be called")
	}
}

func TestImportUploadError(t *testing.T) {
	ts := &testServer{
		loginOK:    true,
		downloadOK: true,
		categories: []Category{
			{ID: "c1", Name: "Fail", NbImages: 1},
		},
		images: map[string][]ImageInfo{
			"c1": {
				{ID: "img1", File: "fail.jpg", Name: "Fail", MD5Sum: "fail1"},
			},
		},
	}
	srv := httptest.NewServer(ts.handler())
	defer srv.Close()

	mock := newMockFlickr()
	mock.failUpload = true
	imp := &Importer{Events: silentEvents(), Flickr: mock}

	summary, err := imp.Import(context.Background(), &ImportOptions{URL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", summary.Failed)
	}
	if summary.Succeeded != 0 {
		t.Errorf("Succeeded = %d, want 0", summary.Succeeded)
	}
}
