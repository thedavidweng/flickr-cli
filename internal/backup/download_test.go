package backup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func TestPlanModeFromFlags(t *testing.T) {
	tests := []struct {
		layout    string
		all       bool
		hasAlbums bool
		want      PlanMode
	}{
		{"id-dirs", false, false, BackupIDDirs},
		{"album", false, false, BackupAlbums},
		{"", true, false, BackupAlbums},
		{"", false, true, BackupAlbums},
		{"", false, false, BackupUser},
	}
	for _, tc := range tests {
		got := PlanModeFromFlags(tc.layout, tc.all, tc.hasAlbums)
		if got != tc.want {
			t.Errorf("PlanModeFromFlags(%q, %v, %v) = %q, want %q",
				tc.layout, tc.all, tc.hasAlbums, got, tc.want)
		}
	}
}

func TestPlanToItemsAlbumMode(t *testing.T) {
	plan := &BackupPlan{
		Items: []BackupItem{
			{PhotoID: "p1", Title: "Sunset", AlbumID: "a1", AlbumName: "Vacation", Media: "photo", OriginalFormat: "jpg"},
			{PhotoID: "p2", Title: "Beach", AlbumID: "a1", AlbumName: "Vacation", Media: "photo", OriginalFormat: "jpg"},
		},
	}
	opts := &BackupPlanOptions{Mode: BackupAlbums, Dest: "/tmp/backup"}

	items := planToItems(plan, opts)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].PhotoID != "p1" {
		t.Errorf("expected p1, got %s", items[0].PhotoID)
	}
	// Album mode should include album name in path
	if items[0].FilePath == "/tmp/backup/Sunset.jpg" {
		t.Errorf("expected album name in path, got %s", items[0].FilePath)
	}
}

func TestPlanToItemsIDDirsMode(t *testing.T) {
	plan := &BackupPlan{
		Items: []BackupItem{
			{PhotoID: "p1", Media: "photo", OriginalFormat: "jpg"},
		},
	}
	opts := &BackupPlanOptions{Mode: BackupIDDirs, Dest: "/tmp/backup"}

	items := planToItems(plan, opts)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	// ID-dirs mode uses hashed path
	if items[0].FilePath == "/tmp/backup/p1.jpg" {
		t.Errorf("expected hashed path for id-dirs mode, got %s", items[0].FilePath)
	}
}

func TestPlanToItemsMetadata(t *testing.T) {
	plan := &BackupPlan{
		Items: []BackupItem{
			{PhotoID: "p1", Title: "Test", Media: "photo", OriginalFormat: "jpg"},
		},
	}
	opts := &BackupPlanOptions{Mode: BackupUser, Dest: "/tmp/backup"}

	items := planToItems(plan, opts)
	if items[0].MetadataPathJSON != "" || items[0].MetadataPathYAML != "" {
		t.Errorf("expected no metadata paths by default")
	}

	opts.Metadata = "json"
	items = planToItems(plan, opts)
	if items[0].MetadataPathJSON == "" {
		t.Error("expected JSON metadata path")
	}

	opts.Metadata = "both"
	items = planToItems(plan, opts)
	if items[0].MetadataPathJSON == "" || items[0].MetadataPathYAML == "" {
		t.Error("expected both metadata paths")
	}
}

func TestIDsToItems(t *testing.T) {
	cfg := &DownloadConfig{Dest: "/tmp/backup", Size: "original"}
	ids := []string{"p1", "p2"}

	items := idsToItems(ids, cfg)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].PhotoID != "p1" {
		t.Errorf("expected p1, got %s", items[0].PhotoID)
	}
	if items[0].SizeLabel != "original" {
		t.Errorf("expected original, got %s", items[0].SizeLabel)
	}
}

func TestIDsToItemsMetadata(t *testing.T) {
	cfg := &DownloadConfig{Dest: "/tmp/backup", Metadata: "both"}
	items := idsToItems([]string{"p1"}, cfg)
	if items[0].MetadataPathJSON == "" || items[0].MetadataPathYAML == "" {
		t.Error("expected both metadata paths")
	}
}

func TestDefaultHTTPClient(t *testing.T) {
	c := DefaultHTTPClient()
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.Timeout <= 0 {
		t.Errorf("expected positive timeout, got %v", c.Timeout)
	}
}

func TestDownloadByIDs(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Test", Owner: "me"}

	// Set up a fake download server
	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("fake-image"))
	}))
	defer dlServer.Close()

	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: dlServer.URL + "/p1.jpg", Width: 100, Height: 100},
	}

	cfg := &DownloadConfig{
		Dest:        t.TempDir(),
		Size:        "original",
		Force:       true,
		Concurrency: 1,
	}

	summary, err := DownloadByIDs(context.Background(), fake.Client(), dlServer.Client(), []string{"p1"}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 1 {
		t.Errorf("expected total 1, got %d", summary.Total)
	}
	if summary.Completed != 1 {
		t.Errorf("expected completed 1, got %d", summary.Completed)
	}
}

func TestDownloadByIDsSkipExisting(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Test", Owner: "me"}

	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fake-image"))
	}))
	defer dlServer.Close()

	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: dlServer.URL + "/p1.jpg", Width: 100, Height: 100},
	}

	dest := t.TempDir()
	cfg := &DownloadConfig{
		Dest:        dest,
		Size:        "original",
		Force:       false, // don't force — should skip
		Concurrency: 1,
	}

	// First download
	summary1, err := DownloadByIDs(context.Background(), fake.Client(), dlServer.Client(), []string{"p1"}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary1.Completed != 1 {
		t.Errorf("expected first download completed=1, got %d", summary1.Completed)
	}

	// Second download should skip
	summary2, err := DownloadByIDs(context.Background(), fake.Client(), dlServer.Client(), []string{"p1"}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary2.Skipped != 1 {
		t.Errorf("expected second download skipped=1, got %d", summary2.Skipped)
	}
}

func TestDownloadByPlanEmpty(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)

	cfg := &DownloadConfig{Dest: t.TempDir(), Concurrency: 1}
	opts := &BackupPlanOptions{Mode: BackupUser, Dest: cfg.Dest}

	summary, err := DownloadByPlan(context.Background(), fake.Client(), http.DefaultClient, opts, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 0 {
		t.Errorf("expected total 0 for empty plan, got %d", summary.Total)
	}
}

func TestDownloadByPlanWithItems(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Test", Owner: "me"}

	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("fake-image"))
	}))
	defer dlServer.Close()

	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: dlServer.URL + "/p1.jpg", Width: 100, Height: 100},
	}

	cfg := &DownloadConfig{
		Dest:        t.TempDir(),
		Size:        "original",
		Force:       true,
		Concurrency: 1,
	}
	opts := &BackupPlanOptions{Mode: BackupUser, Dest: cfg.Dest, Force: true}

	summary, err := DownloadByPlan(context.Background(), fake.Client(), dlServer.Client(), opts, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 1 {
		t.Errorf("expected total 1, got %d", summary.Total)
	}
	if summary.Completed != 1 {
		t.Errorf("expected completed 1, got %d", summary.Completed)
	}
}
