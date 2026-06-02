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
		w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Vacation"},"photos":10}]}}`))
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

	plan, err := BuildPlan(context.Background(), client, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(plan.Items))
	}
}

func TestBuildPlanAlbumsAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","photosets":{"photoset":[{"id":"album-1","title":{"_content":"Vacation"},"photos":10}]}}`))
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

	plan, err := BuildPlan(context.Background(), client, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(plan.Items))
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

	_, err := BuildPlan(context.Background(), client, opts)
	if err == nil {
		t.Error("expected error for no album selection")
	}
}

func TestBuildPlanUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Test"}]}}`))
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

	plan, err := BuildPlan(context.Background(), client, opts)
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
		w.Write([]byte(`{"stat":"ok","photos":{"photo":[{"id":"photo-1","title":"Test"}]}}`))
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

	plan, err := BuildPlan(context.Background(), client, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(plan.Items))
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

	_, err := BuildPlan(context.Background(), client, opts)
	if err == nil {
		t.Error("expected error for invalid mode")
	}
}
