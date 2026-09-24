package backup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

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
