package cli

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/checksum"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func TestChecksumsHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"checksums", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestChecksumsSearchDryRun(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Sunset", Owner: "user1", Tags: "checksum:md5=abc123"}
	fake.Photos["p2"] = testutil.FakePhoto{ID: "p2", Title: "Mountains", Owner: "user2", Tags: "checksum:sha1=def456"}

	cmd, buf := cmdContext(t, cfg, true)
	cmd.Flags().String("user-id", "", "")
	err := checksumsSearchCmd.RunE(cmd, []string{"abc123"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "checksums.search" {
		t.Errorf("expected command=checksums.search, got %s", env.Meta.Command)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("expected data.items to be an array, got %T", data["items"])
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(items))
	}

	// Verify the machine_tags parameter was passed correctly
	call, ok := fake.LastCall("flickr.photos.search")
	if !ok {
		t.Fatal("expected call to flickr.photos.search")
	}
	if call.Params["machine_tags"] != "checksum:*=abc123" {
		t.Errorf("expected machine_tags=checksum:*=abc123, got %s", call.Params["machine_tags"])
	}
}

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
	tmpFile := filepath.Join(tmpDir, "test-download")

	hash, err := downloadAndHash(server.Client(), server.URL+"/photo.jpg", tmpFile, "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, hash)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(data) != string(fixture) {
		t.Errorf("file content mismatch")
	}
}

func TestDownloadAndHashHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-err")

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
	tmpFile := filepath.Join(tmpDir, "test-sha1")

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
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "test-user-123",
		Tags:  "checksum:md5=abc123,nature",
	}

	cmd, _ := cmdContext(t, cfg, false)

	has, err := photoHasChecksum(fake.Client(), cmd, "p1", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected photo to have md5 checksum")
	}
}

func TestPhotoHasChecksumNotFound(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p2"] = testutil.FakePhoto{
		ID:    "p2",
		Title: "No checksum",
		Owner: "test-user-123",
		Tags:  "nature,landscape",
	}

	cmd, _ := cmdContext(t, cfg, false)

	has, err := photoHasChecksum(fake.Client(), cmd, "p2", "md5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected photo to NOT have md5 checksum")
	}
}

// --- getOriginalSourceURL ---

func TestGetOriginalSourceURLFound(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test",
		Owner: "test-user-123",
	}

	cmd, _ := cmdContext(t, cfg, false)

	url, err := getOriginalSourceURL(fake.Client(), cmd, "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty source URL")
	}
}

// --- verifyPhoto integration ---

func TestVerifyPhotoMatching(t *testing.T) {
	fixture := []byte("photo bytes for checksum test")
	expectedHash := fmt.Sprintf("%x", md5.Sum(fixture))

	photoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer photoServer.Close()

	fake, cfg := setupFakeCLI(t)
	machineTag := checksum.FormatMachineTag("md5", expectedHash)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Test Photo",
		Owner: "test-user-123",
		Tags:  machineTag,
	}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: photoServer.URL + "/p1.jpg", Width: 4000, Height: 3000},
	}

	cmd, _ := cmdContext(t, cfg, false)
	result := verifyPhoto(fake.Client(), cmd, "p1", t.TempDir())
	if result.Status != checksum.VerifyValid {
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

	fake, cfg := setupFakeCLI(t)
	machineTag := checksum.FormatMachineTag("md5", wrongHash)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "Mismatch Photo",
		Owner: "test-user-123",
		Tags:  machineTag,
	}
	fake.PhotoSizes["p1"] = []testutil.FakeSize{
		{Label: "Original", Source: photoServer.URL + "/p1.jpg", Width: 4000, Height: 3000},
	}

	cmd, _ := cmdContext(t, cfg, false)
	result := verifyPhoto(fake.Client(), cmd, "p1", t.TempDir())
	if result.Status != checksum.VerifyMismatch {
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
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{
		ID:    "p1",
		Title: "No Checksum",
		Owner: "test-user-123",
		Tags:  "nature",
	}

	cmd, _ := cmdContext(t, cfg, false)
	result := verifyPhoto(fake.Client(), cmd, "p1", t.TempDir())
	if result.Status != checksum.VerifyMissing {
		t.Errorf("expected VerifyMissing, got %v", result.Status)
	}
}

func TestChecksumsAddReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)
	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
	err := checksumsAddCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error with --read-only")
	}
	env := parseEnvelope(t, buf)
	if env.Error.Code != model.ErrReadOnlyViolation {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}
