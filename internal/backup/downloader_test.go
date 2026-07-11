package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// mockFlickrAPI is a controllable test double for flickr.FlickrAPI.
type mockFlickrAPI struct {
	sizes        []flickr.Size
	sizesErr     error
	videoStreams []flickr.VideoStream
	videoErr     error
	exifData     *flickr.ExifData
	exifErr      error
	callErr      error
	callHandler  func(method string, params map[string]string) (json.RawMessage, error)
}

func (m *mockFlickrAPI) Call(_ context.Context, method string, params map[string]string, out any) error {
	if m.callErr != nil {
		return m.callErr
	}
	if m.callHandler != nil {
		raw, err := m.callHandler(method, params)
		if err != nil {
			return err
		}
		if out != nil && raw != nil {
			return json.Unmarshal(raw, out)
		}
		return nil
	}
	return nil
}
func (m *mockFlickrAPI) CallRaw(_ context.Context, _ string, _ map[string]string) (json.RawMessage, error) {
	return nil, nil
}
func (m *mockFlickrAPI) TestLogin(_ context.Context) (*flickr.LoginInfo, error) {
	return nil, nil
}
func (m *mockFlickrAPI) TestEcho(_ context.Context) error { return nil }
func (m *mockFlickrAPI) GetSizes(_ context.Context, _ string) ([]flickr.Size, error) {
	return m.sizes, m.sizesErr
}
func (m *mockFlickrAPI) GetVideoStreams(_ context.Context, _ string) ([]flickr.VideoStream, error) {
	return m.videoStreams, m.videoErr
}
func (m *mockFlickrAPI) GetExif(_ context.Context, _ string) (*flickr.ExifData, error) {
	return m.exifData, m.exifErr
}
func (m *mockFlickrAPI) Upload(_ context.Context, _ string, _ *flickr.UploadOptions) (*flickr.UploadResult, error) {
	return nil, nil
}
func (m *mockFlickrAPI) AddToAlbum(_ context.Context, _, _ string) error { return nil }
func (m *mockFlickrAPI) IsAuthenticated() bool                           { return true }
func (m *mockFlickrAPI) RequestToken(_ context.Context, _ string) (*flickr.RequestTokenResponse, error) {
	return nil, nil
}
func (m *mockFlickrAPI) AuthorizationURL(_, _ string) string { return "" }
func (m *mockFlickrAPI) AccessToken(_ context.Context, _, _, _ string) (*flickr.AccessTokenResponse, error) {
	return nil, nil
}
func (m *mockFlickrAPI) GetMethods(_ context.Context) (json.RawMessage, error) { return nil, nil }
func (m *mockFlickrAPI) GetMethodInfo(_ context.Context, _ string) (json.RawMessage, error) {
	return nil, nil
}

var _ flickr.FlickrAPI = (*mockFlickrAPI)(nil)

func TestDownloaderCreation(t *testing.T) {
	downloader := &Downloader{
		Concurrency: 4,
		Events:      &output.EventWriter{},
	}

	if downloader.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", downloader.Concurrency)
	}
}

func TestDownloadOptions(t *testing.T) {
	opts := DownloadOptions{
		Force:    true,
		Size:     "original",
		Metadata: "json",
	}

	if !opts.Force {
		t.Error("expected force=true")
	}
	if opts.Size != "original" {
		t.Errorf("expected original, got %s", opts.Size)
	}
}

func TestDownloadSummary(t *testing.T) {
	summary := &DownloadSummary{
		Total:     10,
		Completed: 8,
		Skipped:   1,
		Failed:    1,
	}

	if summary.Total != 10 {
		t.Errorf("expected total 10, got %d", summary.Total)
	}
	if summary.Completed != 8 {
		t.Errorf("expected completed 8, got %d", summary.Completed)
	}
}

func TestDownloadItem(t *testing.T) {
	item := DownloadItem{
		PhotoID:          "photo-123",
		FilePath:         "/tmp/photo.jpg",
		MetadataPathJSON: "/tmp/photo.json",
		SizeLabel:        "original",
	}

	if item.PhotoID != "photo-123" {
		t.Errorf("expected photo-123, got %s", item.PhotoID)
	}
}

func TestDownload(t *testing.T) {
	// Create a mock server that returns a photo
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","sizes":{"size":[{"label":"Original","width":4000,"height":3000,"source":"` + photoServer.URL + `/photo.jpg","url":"` + photoServer.URL + `/photo.jpg","media":"photo"}]}}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      client,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "photo-123", FilePath: filePath, SizeLabel: "original"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", summary.Completed)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); err != nil {
		t.Error("downloaded file should exist")
	}
}

func TestDownloadSkipExisting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	downloader := &Downloader{
		HTTP:        http.DefaultClient,
		Client:      &flickr.Client{},
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "photo-123", FilePath: filePath, SizeLabel: "original"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{Force: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", summary.Skipped)
	}
	if summary.Completed != 0 {
		t.Errorf("expected 0 completed, got %d", summary.Completed)
	}
}

func TestDownloadConcurrency(t *testing.T) {
	// Track max concurrent downloads
	var concurrent int32
	var maxConcurrent int32

	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&concurrent, 1)
		// Update max
		for {
			old := atomic.LoadInt32(&maxConcurrent)
			if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&concurrent, -1)
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","sizes":{"size":[{"label":"Original","width":100,"height":100,"source":"` + photoServer.URL + `/photo.jpg","url":"` + photoServer.URL + `/photo.jpg","media":"photo"}]}}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	// Download 4 photos with concurrency 4
	var mu sync.Mutex
	var paths []string
	for i := 0; i < 4; i++ {
		p := filepath.Join(tmpDir, "photo"+string(rune('A'+i))+".jpg")
		mu.Lock()
		paths = append(paths, p)
		mu.Unlock()
	}

	items := make([]DownloadItem, 4)
	for i, p := range paths {
		items[i] = DownloadItem{PhotoID: "photo-" + string(rune('A'+i)), FilePath: p, SizeLabel: "original"}
	}

	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      client,
		Concurrency: 4,
		Events:      &output.EventWriter{},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 4 {
		t.Errorf("expected 4 completed, got %d", summary.Completed)
	}

	peak := atomic.LoadInt32(&maxConcurrent)
	if peak < 2 {
		t.Errorf("expected concurrent downloads (peak=%d), but downloads ran sequentially", peak)
	}
}

func TestReplaceExt(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		newExt   string
		want     string
	}{
		{
			name:     "simple extension replacement",
			filePath: filepath.Join("photos", "image.jpg"),
			newExt:   "png",
			want:     filepath.Join("photos", "image.png"),
		},
		{
			name:     "deeply nested path",
			filePath: filepath.Join("backup", "ab", "cd", "12345", "12345.jpg"),
			newExt:   "png",
			want:     filepath.Join("backup", "ab", "cd", "12345", "12345.png"),
		},
		{
			name:     "no existing extension",
			filePath: filepath.Join("photos", "image"),
			newExt:   "jpg",
			want:     filepath.Join("photos", "image") + ".jpg",
		},
		{
			name:     "double extension in filename",
			filePath: filepath.Join("photos", "my.photo.jpg"),
			newExt:   "png",
			want:     filepath.Join("photos", "my.photo.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceExt(tt.filePath, tt.newExt)
			if got != tt.want {
				t.Errorf("replaceExt(%q, %q) = %q, want %q", tt.filePath, tt.newExt, got, tt.want)
			}
		})
	}
}

func TestDownloadVideoStreamErrorFallback(t *testing.T) {
	// When GetVideoStreams fails, the downloader should log the error
	// and fall back to GetSizes.
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		videoErr: fmt.Errorf("video API unavailable"),
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "vid-1", FilePath: filepath.Join(tmpDir, "vid.mp4"), Media: "video"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed (fallback to getSizes), got %d", summary.Completed)
	}
}

func TestDownloadVideoStreamSelectBestError(t *testing.T) {
	// When SelectBestStream fails (empty streams after filtering),
	// the downloader should log and fall back to GetSizes.
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		videoStreams: []flickr.VideoStream{}, // empty → SelectBestStream returns error
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "vid-2", FilePath: filepath.Join(tmpDir, "vid.mp4"), Media: "video"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed (fallback to getSizes), got %d", summary.Completed)
	}
}

func TestDownloadVideoStreamSuccess(t *testing.T) {
	// When GetVideoStreams succeeds, the downloader should use the video URL.
	videoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-video-data"))
	}))
	defer videoServer.Close()

	mock := &mockFlickrAPI{
		videoStreams: []flickr.VideoStream{
			{Type: "orig", Width: 1920, Height: 1080, Source: videoServer.URL + "/video.mp4"},
		},
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        videoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "vid-3", FilePath: filepath.Join(tmpDir, "vid.mp4"), Media: "video"},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", summary.Completed)
	}
}

func TestDownloadGetSizesError(t *testing.T) {
	mock := &mockFlickrAPI{
		sizesErr: fmt.Errorf("API error"),
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        http.DefaultClient,
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "p-err", FilePath: filepath.Join(tmpDir, "photo.jpg")},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", summary.Failed)
	}
}

func TestDownloadContextCancellation(t *testing.T) {
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = w.Write([]byte("slow"))
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so the request fails immediately

	items := []DownloadItem{
		{PhotoID: "p-cancel", FilePath: filepath.Join(tmpDir, "photo.jpg")},
	}

	summary, _ := downloader.Download(ctx, items, DownloadOptions{})
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed for canceled context, got %d", summary.Failed)
	}
}

func TestDownloadWithMetadataSidecarError(t *testing.T) {
	// When writeSidecars encounters a getInfo error, the download should
	// still succeed (sidecar errors are logged, not fatal).
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
		callErr: fmt.Errorf("getInfo failed"),
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{
			PhotoID:          "p-meta",
			FilePath:         filePath,
			MetadataPathJSON: filePath + ".json",
		},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{Exif: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed (sidecar error is non-fatal), got %d", summary.Completed)
	}
}

func TestDownloadWithMetadataSidecarSuccess(t *testing.T) {
	// When writeSidecars succeeds, both JSON and YAML sidecar files
	// should be written alongside the downloaded photo.
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
		callHandler: func(method string, _ map[string]string) (json.RawMessage, error) {
			if method == "flickr.photos.getInfo" {
				return json.RawMessage(`{"stat":"ok","photo":{"id":"p-ok","title":{"_content":"My Photo"}}}`), nil
			}
			return nil, nil
		},
		exifData: &flickr.ExifData{PhotoID: "p-ok", Tags: []flickr.ExifTag{{Tag: "Make", Raw: "Canon"}}},
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{
			PhotoID:          "p-ok",
			FilePath:         filePath,
			MetadataPathJSON: filePath + ".json",
			MetadataPathYAML: filePath + ".yaml",
		},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{Exif: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Fatalf("expected 1 completed, got %d", summary.Completed)
	}

	// Verify JSON sidecar was written
	jsonData, err := os.ReadFile(filePath + ".json")
	if err != nil {
		t.Fatalf("expected JSON sidecar to exist: %v", err)
	}
	var jsonMap map[string]any
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Fatalf("JSON sidecar is invalid: %v", err)
	}

	// Verify YAML sidecar was written
	yamlData, err := os.ReadFile(filePath + ".yaml")
	if err != nil {
		t.Fatalf("expected YAML sidecar to exist: %v", err)
	}
	var yamlMap map[string]any
	if err := yaml.Unmarshal(yamlData, &yamlMap); err != nil {
		t.Fatalf("YAML sidecar is invalid: %v", err)
	}
}

func TestDownloadWithMetadataSidecarExifError(t *testing.T) {
	// When GetExif fails but getInfo succeeds, the download should
	// still succeed and sidecars should still be written (without exif).
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-photo-data"))
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
		callHandler: func(method string, _ map[string]string) (json.RawMessage, error) {
			if method == "flickr.photos.getInfo" {
				return json.RawMessage(`{"stat":"ok","photo":{"id":"p-exif-err","title":{"_content":"My Photo"}}}`), nil
			}
			return nil, nil
		},
		exifErr: fmt.Errorf("exif API unavailable"),
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "photo.jpg")
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{
			PhotoID:          "p-exif-err",
			FilePath:         filePath,
			MetadataPathJSON: filePath + ".json",
		},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{Exif: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", summary.Completed)
	}
	// JSON sidecar should still exist (without exif field)
	if _, err := os.Stat(filePath + ".json"); err != nil {
		t.Error("expected JSON sidecar to exist despite exif error")
	}
}

func TestDownloadSelectSizeError(t *testing.T) {
	// When SelectSize fails (empty sizes), the download should fail.
	mock := &mockFlickrAPI{
		sizes: []flickr.Size{}, // empty → SelectSize returns error
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        http.DefaultClient,
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "p-size-err", FilePath: filepath.Join(tmpDir, "photo.jpg")},
	}

	summary, err := downloader.Download(context.Background(), items, DownloadOptions{Size: "original"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", summary.Failed)
	}
}

func TestDownloadHTTPNon200(t *testing.T) {
	// When the download URL returns a non-200 status, the item should fail.
	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer photoServer.Close()

	mock := &mockFlickrAPI{
		sizes: []flickr.Size{
			{Label: "Original", Width: 4000, Height: 3000, Source: photoServer.URL + "/photo.jpg", Media: "photo"},
		},
	}

	tmpDir := t.TempDir()
	downloader := &Downloader{
		HTTP:        photoServer.Client(),
		Client:      mock,
		Concurrency: 1,
		Events:      &output.EventWriter{},
	}

	items := []DownloadItem{
		{PhotoID: "p-403", FilePath: filepath.Join(tmpDir, "photo.jpg")},
	}

	summary, _ := downloader.Download(context.Background(), items, DownloadOptions{})
	if summary.Failed != 1 {
		t.Errorf("expected 1 failed for HTTP 403, got %d", summary.Failed)
	}
}
