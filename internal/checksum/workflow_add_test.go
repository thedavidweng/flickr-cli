package checksum

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func md5Hex(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func TestTaggerAddDryRun(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Test", Owner: "me"}
	fake.Photos["p2"] = testutil.FakePhoto{ID: "p2", Title: "Test2", Owner: "me"}

	tagger := &Tagger{API: fake.Client(), HTTP: http.DefaultClient}

	result, err := tagger.Add(context.Background(), AddOptions{
		HashAlgo: "md5",
		UserID:   "me",
		DryRun:   true,
		PerPage:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Planned {
		t.Error("expected planned=true for dry run")
	}
	if len(result.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(result.Details))
	}
	for _, d := range result.Details {
		if d.Status != "would_add" {
			t.Errorf("expected would_add, got %s", d.Status)
		}
	}
}

func TestTaggerAddInvalidAlgorithm(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	tagger := &Tagger{API: fake.Client(), HTTP: http.DefaultClient}

	_, err := tagger.Add(context.Background(), AddOptions{
		HashAlgo: "invalid",
		UserID:   "me",
	})
	if err == nil {
		t.Fatal("expected error for invalid algorithm")
	}
}

func TestTaggerAddSkipExisting(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "me",
		Tags:  "checksum:md5=abc123",
	}

	tagger := &Tagger{API: fake.Client(), HTTP: http.DefaultClient}

	result, err := tagger.Add(context.Background(), AddOptions{
		HashAlgo: "md5",
		UserID:   "me",
		PerPage:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("expected skipped=1, got %d", result.Skipped)
	}
}

func TestTaggerAddSuccess(t *testing.T) {
	fixture := []byte("test image data")
	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer dlServer.Close()

	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Test", Owner: "me"}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: dlServer.URL + "/p1.jpg", Width: 100, Height: 100},
	}

	tagger := &Tagger{API: fake.Client(), HTTP: dlServer.Client()}

	result, err := tagger.Add(context.Background(), AddOptions{
		HashAlgo: "md5",
		UserID:   "me",
		PerPage:  50,
		TmpDir:   t.TempDir(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Added != 1 {
		t.Errorf("expected added=1, got %d (details: %+v)", result.Added, result.Details)
	}
}

func TestVerifierVerify(t *testing.T) {
	fixture := []byte("photo bytes for verify")
	expectedHash := md5Hex(fixture)

	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer dlServer.Close()

	fake := testutil.NewFakeFlickr(t)
	machineTag := FormatMachineTag("md5", expectedHash)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "me",
		Tags:  machineTag,
	}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: dlServer.URL + "/p1.jpg", Width: 100, Height: 100},
	}

	verifier := &Verifier{API: fake.Client(), HTTP: dlServer.Client()}

	report, err := verifier.Verify(context.Background(), VerifyOptions{
		TmpDir:  t.TempDir(),
		PerPage: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Valid != 1 {
		t.Errorf("expected valid=1, got %d (results: %+v)", report.Summary.Valid, report.Results)
	}
}

func TestVerifierVerifyMismatch(t *testing.T) {
	fixture := []byte("actual bytes")
	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer dlServer.Close()

	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "me",
		Tags:  FormatMachineTag("md5", "0000000000000000000000000000dead"),
	}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: dlServer.URL + "/p1.jpg", Width: 100, Height: 100},
	}

	verifier := &Verifier{API: fake.Client(), HTTP: dlServer.Client()}

	report, err := verifier.Verify(context.Background(), VerifyOptions{
		TmpDir:  t.TempDir(),
		PerPage: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Mismatch != 1 {
		t.Errorf("expected mismatch=1, got %d", report.Summary.Mismatch)
	}
}
