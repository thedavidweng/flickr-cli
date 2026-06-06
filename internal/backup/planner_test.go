package backup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

func TestMatchAlbumTitle(t *testing.T) {
	tests := []struct {
		album   string
		pattern string
		match   bool
	}{
		{"Vacation", "Vacation", true},
		{"vacation", "Vacation", true},
		{"Vacation 2024", "Vacation*", true},
		{"My Album", "Other", false},
	}

	for _, tt := range tests {
		t.Run(tt.album+"_"+tt.pattern, func(t *testing.T) {
			got := matchAlbumTitle(tt.album, tt.pattern)
			if got != tt.match {
				t.Errorf("matchAlbumTitle(%q, %q) = %v, want %v", tt.album, tt.pattern, got, tt.match)
			}
		})
	}
}

func TestPlanModes(t *testing.T) {
	if BackupAlbums != "albums" {
		t.Errorf("expected albums, got %s", BackupAlbums)
	}
	if BackupUser != "user" {
		t.Errorf("expected user, got %s", BackupUser)
	}
	if BackupIDDirs != "id_dirs" {
		t.Errorf("expected id_dirs, got %s", BackupIDDirs)
	}
}

func TestBackupPlanOptions(t *testing.T) {
	opts := BackupPlanOptions{
		Mode:        BackupAlbums,
		Dest:        "/tmp/backup",
		AlbumTitles: []string{"Vacation"},
		All:         false,
		Size:        "original",
		Metadata:    "json",
	}

	if opts.Mode != BackupAlbums {
		t.Errorf("expected albums mode, got %s", opts.Mode)
	}
	if opts.Dest != "/tmp/backup" {
		t.Errorf("expected /tmp/backup, got %s", opts.Dest)
	}
}

func TestBackupItem(t *testing.T) {
	item := BackupItem{
		PhotoID: "photo-123",
		Title:   "My Photo",
		AlbumID: "album-456",
	}

	if item.PhotoID != "photo-123" {
		t.Errorf("expected photo-123, got %s", item.PhotoID)
	}
	if item.Title != "My Photo" {
		t.Errorf("expected My Photo, got %s", item.Title)
	}
}

func TestBuildPlanAlbums(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.FormValue("method") {
		case "flickr.photosets.getList":
			_, _ = w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Vacation"},"photos":2}]}}`))
		case "flickr.photosets.getPhotos":
			_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"id":"album-1","photo":[{"id":"photo-1","title":"Sunset"},{"id":"photo-2","title":"Beach"}],"page":1,"pages":1,"perpage":100,"total":2}}`))
		default:
			_, _ = w.Write([]byte(`{"stat":"ok"}`))
		}
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	opts := BackupPlanOptions{
		Mode:        BackupAlbums,
		AlbumTitles: []string{"Vacation"},
		Size:        "original",
	}

	plan, err := BuildPlan(context.Background(), client, &opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(plan.Items))
	}
	// Verify actual photo IDs, not the album ID.
	ids := map[string]bool{}
	for _, item := range plan.Items {
		if item.PhotoID == "album-1" {
			t.Errorf("PhotoID must be the photo ID, not the album ID; got %q", item.PhotoID)
		}
		if item.AlbumID != "album-1" {
			t.Errorf("expected AlbumID album-1, got %q", item.AlbumID)
		}
		ids[item.PhotoID] = true
	}
	if !ids["photo-1"] || !ids["photo-2"] {
		t.Errorf("expected photo-1 and photo-2 in plan, got %v", ids)
	}
}

func TestBuildPlanAlbumsAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.FormValue("method") {
		case "flickr.photosets.getList":
			_, _ = w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Vacation"},"photos":2}]}}`))
		case "flickr.photosets.getPhotos":
			_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"id":"album-1","photo":[{"id":"photo-1","title":"Sunset"},{"id":"photo-2","title":"Beach"}],"page":1,"pages":1,"perpage":100,"total":2}}`))
		default:
			_, _ = w.Write([]byte(`{"stat":"ok"}`))
		}
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	opts := BackupPlanOptions{
		Mode: BackupAlbums,
		All:  true,
		Size: "original",
	}

	plan, err := BuildPlan(context.Background(), client, &opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(plan.Items))
	}
	for _, item := range plan.Items {
		if item.PhotoID == "album-1" {
			t.Errorf("PhotoID must be the photo ID, not the album ID; got %q", item.PhotoID)
		}
		if item.AlbumID != "album-1" {
			t.Errorf("expected AlbumID album-1, got %q", item.AlbumID)
		}
	}
}

func TestBuildPlanAlbumsNoSelection(t *testing.T) {
	client := &flickr.Client{
		APIKey: "test-key",
		HTTP:   http.DefaultClient,
	}

	opts := BackupPlanOptions{
		Mode: BackupAlbums,
		Size: "original",
	}

	_, err := BuildPlan(context.Background(), client, &opts)
	if err == nil {
		t.Error("expected error for no album selection")
	}
}

func TestBuildPlanUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Test"}]}}`))
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	opts := BackupPlanOptions{
		Mode:   BackupUser,
		UserID: "me",
		Size:   "original",
	}

	plan, err := BuildPlan(context.Background(), client, &opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(plan.Items))
	}
}

func TestBuildPlanIDDirs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Test"}]}}`))
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	opts := BackupPlanOptions{
		Mode: BackupIDDirs,
		Size: "original",
	}

	plan, err := BuildPlan(context.Background(), client, &opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(plan.Items))
	}
}

func TestBuildPlanAlbumsPagination(t *testing.T) {
	getListCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.FormValue("method") {
		case "flickr.photosets.getList":
			getListCalls++
			if getListCalls == 1 {
				// Page 1 of 2: 2 albums
	_, _ = w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"a1","title":{"_content":"Album 1"},"photos":5},{"id":"a2","title":{"_content":"Album 2"},"photos":3}],"page":1,"pages":2,"perpage":2,"total":3}}`))
			} else {
				// Page 2 of 2: 1 album
	_, _ = w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"a3","title":{"_content":"Album 3"},"photos":7}],"page":2,"pages":2,"perpage":2,"total":3}}`))
			}
		case "flickr.photosets.getPhotos":
			albumID := r.FormValue("photoset_id")
			switch albumID {
			case "a1":
	_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"id":"a1","photo":[{"id":"p1","title":"P1"},{"id":"p2","title":"P2"}],"page":1,"pages":1,"perpage":100,"total":2}}`))
			case "a2":
	_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"id":"a2","photo":[{"id":"p3","title":"P3"}],"page":1,"pages":1,"perpage":100,"total":1}}`))
			case "a3":
	_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"id":"a3","photo":[{"id":"p4","title":"P4"},{"id":"p5","title":"P5"}],"page":1,"pages":1,"perpage":100,"total":2}}`))
			default:
	_, _ = w.Write([]byte(`{"stat":"ok","photoset":{"photo":[],"page":1,"pages":1,"perpage":100,"total":0}}`))
			}
		default:
			_, _ = w.Write([]byte(`{"stat":"ok"}`))
		}
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	opts := BackupPlanOptions{
		Mode: BackupAlbums,
		All:  true,
		Size: "original",
	}

	plan, err := BuildPlan(context.Background(), client, &opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 albums with 2+1+2=5 total photos
	if len(plan.Items) != 5 {
		t.Errorf("expected 5 items across 3 albums, got %d", len(plan.Items))
	}
	// Verify no item has an album ID as its photo ID.
	for _, item := range plan.Items {
		if item.PhotoID == "a1" || item.PhotoID == "a2" || item.PhotoID == "a3" {
			t.Errorf("PhotoID should be a photo ID, not album ID %q", item.PhotoID)
		}
	}
	if getListCalls != 2 {
		t.Errorf("expected 2 getList calls (one per page), got %d", getListCalls)
	}
}

func TestBuildPlanUserPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
	_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"p1","title":"Photo 1"},{"id":"p2","title":"Photo 2"}],"page":1,"pages":2,"perpage":2,"total":3}}`))
		} else {
	_, _ = w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"p3","title":"Photo 3"}],"page":2,"pages":2,"perpage":2,"total":3}}`))
		}
	}))
	defer server.Close()

	client := &flickr.Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: flickr.Endpoints{REST: server.URL + "/"},
	}

	opts := BackupPlanOptions{
		Mode:   BackupUser,
		UserID: "me",
		Size:   "original",
	}

	plan, err := BuildPlan(context.Background(), client, &opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 3 {
		t.Errorf("expected 3 items across 2 pages, got %d", len(plan.Items))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (one per page), got %d", callCount)
	}
}

func TestBuildPlanInvalidMode(t *testing.T) {
	client := &flickr.Client{
		APIKey: "test-key",
		HTTP:   http.DefaultClient,
	}

	opts := BackupPlanOptions{
		Mode: "invalid",
		Size: "original",
	}

	_, err := BuildPlan(context.Background(), client, &opts)
	if err == nil {
		t.Error("expected error for invalid mode")
	}
}
