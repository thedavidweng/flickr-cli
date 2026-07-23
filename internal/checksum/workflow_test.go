package checksum

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

// --- downloadAndHash ---

func TestDownloadAndHash(t *testing.T) {
	fixture := []byte("hello world")
	expectedHash := fmt.Sprintf("%x", md5.Sum(fixture))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test-download"

	hash, err := downloadAndHash(server.Client(), server.URL+"/photo.jpg", tmpFile, "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, hash)
	}
}

func TestDownloadAndHashHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test-err"

	_, err := downloadAndHash(server.Client(), server.URL+"/photo.jpg", tmpFile, "md5")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestDownloadAndHashSHA1(t *testing.T) {
	fixture := []byte("test sha1 content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test-sha1"

	hash, err := downloadAndHash(server.Client(), server.URL+"/photo.jpg", tmpFile, "sha1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 40 {
		t.Errorf("expected 40 char SHA1 hash, got %d chars: %s", len(hash), hash)
	}
}

// --- photoHasChecksum ---

func TestPhotoHasChecksumFound(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "test-user-123",
		Tags:  "checksum:md5=abc123,nature",
	}

	has, err := photoHasChecksum(context.Background(), fake.Client(), "p1", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected photo to have md5 checksum")
	}
}

func TestPhotoHasChecksumNotFound(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p2"] = testutil.FakePhoto{
		ID:    "p2",
		Title: "No checksum",
		Owner: "test-user-123",
		Tags:  "nature,landscape",
	}

	has, err := photoHasChecksum(context.Background(), fake.Client(), "p2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected photo to NOT have md5 checksum")
	}
}

// --- originalSourceURL ---

func TestOriginalSourceURLFound(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "test-user-123",
	}

	url, err := originalSourceURL(context.Background(), fake.Client(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty source URL")
	}
}

// --- verifyPhoto ---

func TestVerifyPhotoMatching(t *testing.T) {
	fixture := []byte("photo bytes for checksum test")
	expectedHash := fmt.Sprintf("%x", md5.Sum(fixture))

	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer photoServer.Close()

	fake := testutil.NewFakeFlickr(t)
	machineTag := FormatMachineTag("md5", expectedHash)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test Photo",
		Owner: "test-user-123",
		Tags:  machineTag,
	}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: photoServer.URL + "/p1.jpg", Width: 4000, Height: 3000},
	}

	result := verifyPhoto(context.Background(), fake.Client(), photoServer.Client(), "p1", t.TempDir())
	if result.Status != VerifyValid {
		t.Errorf("expected VerifyValid, got %v (error: %s)", result.Status, result.Error)
	}
	if result.Expected != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, result.Expected)
	}
}

func TestVerifyPhotoMismatch(t *testing.T) {
	fixture := []byte("actual photo bytes")
	actualHash := fmt.Sprintf("%x", md5.Sum(fixture))
	wrongHash := "0000000000000000000000000000dead"

	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer photoServer.Close()

	fake := testutil.NewFakeFlickr(t)
	machineTag := FormatMachineTag("md5", wrongHash)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Mismatch Photo",
		Owner: "test-user-123",
		Tags:  machineTag,
	}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: photoServer.URL + "/p1.jpg", Width: 4000, Height: 3000},
	}

	result := verifyPhoto(context.Background(), fake.Client(), photoServer.Client(), "p1", t.TempDir())
	if result.Status != VerifyMismatch {
		t.Errorf("expected VerifyMismatch, got %v", result.Status)
	}
	if result.Expected != wrongHash {
		t.Errorf("expected %s, got %s", wrongHash, result.Expected)
	}
	if result.Actual != actualHash {
		t.Errorf("expected actual %s, got %s", actualHash, result.Actual)
	}
}

func TestVerifyPhotoMissingChecksum(t *testing.T) {
	fake := testutil.NewFakeFlickr(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "No Checksum",
		Owner: "test-user-123",
		Tags:  "nature",
	}

	result := verifyPhoto(context.Background(), fake.Client(), http.DefaultClient, "p1", t.TempDir())
	if result.Status != VerifyMissing {
		t.Errorf("expected VerifyMissing, got %v", result.Status)
	}
}

// Ensure bytes import is used (for potential future assertions).
var _ = bytes.Equal
