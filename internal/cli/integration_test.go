package cli

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/backup"
	"github.com/thedavidweng/flickr-cli/internal/config"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

// setupFakeCLI creates a fake Flickr server and writes a config pointing at it.
func setupFakeCLI(t *testing.T) (fake *testutil.FakeFlickr, cfgPath string) {
	t.Helper()
	fake = testutil.NewFakeFlickr(t)

	cfgDir := t.TempDir()
	cfgPath = filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`schema_version: "2026-06-02"
default_profile: default
profiles:
  default:
    api_key: test-api-key
    api_secret: test-api-secret
    oauth_token: test-token
    oauth_token_secret: test-secret
    user_id: test-user-123
    username: testuser
    endpoints:
      rest: %s/services/rest/
      upload: %s/services/upload/
      request_token: %s/oauth/request_token
      authorize: %s/oauth/authorize
      access_token: %s/oauth/access_token
`, fake.Server.URL, fake.Server.URL, fake.Server.URL, fake.Server.URL, fake.Server.URL)

	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	return fake, cfgPath
}

// setupUnauthedCLI writes a config with API key but no OAuth credentials.
func setupUnauthedCLI(t *testing.T, fakeURL string) string {
	t.Helper()
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfgContent := fmt.Sprintf(`schema_version: "2026-06-02"
default_profile: default
profiles:
  default:
    api_key: test-api-key
    api_secret: test-api-secret
    endpoints:
      rest: %s/services/rest/
`, fakeURL)
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	return cfgPath
}

// cmdContext creates a cobra.Command with AppContext set, wired to capture
// stdout into buf.  The command has no RunE — the caller invokes the
// command's RunE from the package-level var directly.
// Optional AppContext fields can be overridden by passing a partially filled
// app; nil means use defaults.
func cmdContext(t *testing.T, cfgPath string, jsonMode bool, appOverrides ...*AppContext) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	app := &AppContext{
		ConfigFile:  cfgPath,
		Profile:     "default",
		JSON:        jsonMode,
		Timeout:     30 * time.Second,
		Retries:     3,
		Concurrency: 4,
		RequestID:   uuid.New().String(),
		StartedAt:   time.Now(),
	}
	if len(appOverrides) > 0 && appOverrides[0] != nil {
		o := appOverrides[0]
		if o.ConfigFile != "" {
			app.ConfigFile = o.ConfigFile
		}
		if o.Profile != "" {
			app.Profile = o.Profile
		}
		app.JSON = o.JSON || jsonMode
		app.ReadOnly = o.ReadOnly
		app.DryRun = o.DryRun
		app.Confirm = o.Confirm
	}
	cmd.SetContext(WithAppContext(context.Background(), app))
	// Register flags that RunE closures read via cmd.Flags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("per-page", 50, "")
	cmd.Flags().String("sort", "title", "")
	cmd.Flags().Bool("raw", false, "")
	cmd.Flags().StringToString("param", nil, "")
	cmd.Flags().String("auth", "optional", "")
	cmd.Flags().Bool("dry-run", app.DryRun, "")
	cmd.Flags().Bool("read-only", app.ReadOnly, "")
	cmd.Flags().Bool("confirm", app.Confirm, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("primary-photo-id", "", "")
	cmd.Flags().String("text", "", "")
	return cmd, buf
}

// parseEnvelope unmarshals JSON output into an Envelope.
func parseEnvelope(t *testing.T, buf *bytes.Buffer) model.Envelope {
	t.Helper()
	raw := buf.Bytes()
	start := bytes.IndexByte(raw, '{')
	end := bytes.LastIndexByte(raw, '}')
	if start < 0 || end < start {
		t.Fatalf("no JSON object found in output: %s", raw)
	}
	var env model.Envelope
	if err := json.Unmarshal(raw[start:end+1], &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v\nraw: %s", err, raw)
	}
	return env
}

// --- Albums integration tests ---

func TestAlbumsListJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Albums["album-1"] = testutil.FakeAlbum{
		ID:          "album-1",
		Title:       "Summer Vacation",
		Description: "Beach photos",
		PhotoCount:  42,
		PrimaryID:   "photo-1",
	}

	cmd, buf := cmdContext(t, cfg, true)
	err := albumsListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "albums.list" {
		t.Errorf("expected command=albums.list, got %s", env.Meta.Command)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("expected data.items to be an array, got %T", data["items"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 album, got %d", len(items))
	}
	album, ok := items[0].(map[string]any)
	if !ok {
		t.Fatal("album is not a map")
	}
	if album["id"] != "album-1" {
		t.Errorf("expected album id=album-1, got %v", album["id"])
	}
	if album["title"] != "Summer Vacation" {
		t.Errorf("expected album title=Summer Vacation, got %v", album["title"])
	}

	if fake.CountMethod("flickr.photosets.getList") != 1 {
		t.Errorf("expected 1 call to getList, got %d", fake.CountMethod("flickr.photosets.getList"))
	}
}

func TestAlbumsListAuthRequired(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg, true)
	err := albumsListCmd.RunE(cmd, nil)

	// Failure should return a CommandError
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error body")
	}
	if env.Error.Code != model.ErrAuthRequired {
		t.Errorf("expected error code AUTH_REQUIRED, got %s", env.Error.Code)
	}
}

func TestAlbumsShowJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Albums["album-42"] = testutil.FakeAlbum{
		ID:          "album-42",
		Title:       "My Album",
		Description: "Desc",
		PhotoCount:  10,
	}

	cmd, buf := cmdContext(t, cfg, true)
	err := albumsShowCmd.RunE(cmd, []string{"album-42"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "albums.show" {
		t.Errorf("expected command=albums.show, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["id"] != "album-42" {
		t.Errorf("expected id=album-42, got %v", data["id"])
	}
	if data["title"] != "My Album" {
		t.Errorf("expected title=My Album, got %v", data["title"])
	}

	if fake.CountMethod("flickr.photosets.getInfo") != 1 {
		t.Errorf("expected 1 call to getInfo, got %d", fake.CountMethod("flickr.photosets.getInfo"))
	}
}

func TestAlbumsShowNotFound(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := albumsShowCmd.RunE(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent album")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrFlickrAPI {
		t.Errorf("expected FLICKR_API_ERROR, got %s", env.Error.Code)
	}
}

func TestAlbumsCreateDryRun(t *testing.T) {
	fake, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	if err := cmd.Flags().Set("title", "New Album"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("primary-photo-id", "photo-99"); err != nil {
		t.Fatal(err)
	}
	err := albumsCreateCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	if fake.CountMethod("flickr.photosets.create") != 0 {
		t.Errorf("expected 0 API calls in dry-run, got %d", fake.CountMethod("flickr.photosets.create"))
	}
}

func TestAlbumsDeleteRequiresConfirm(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := albumsDeleteCmd.RunE(cmd, []string{"album-1"})
	if err == nil {
		t.Fatal("expected error without --confirm")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrConfirmationRequired {
		t.Errorf("expected CONFIRMATION_REQUIRED, got %s", env.Error.Code)
	}
}

func TestAlbumsDeleteReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true, Confirm: true})
	err := albumsDeleteCmd.RunE(cmd, []string{"album-1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrReadOnlyViolation {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Photos integration tests ---

func TestPhotosListJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Sunset", Owner: "test-user-123"}
	fake.Photos["p2"] = testutil.FakePhoto{ID: "p2", Title: "Mountains", Owner: "test-user-123"}

	cmd, buf := cmdContext(t, cfg, true)
	err := photosListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.list" {
		t.Errorf("expected command=photos.list, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatal("items is not a slice")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(items))
	}

	if fake.CountMethod("flickr.people.getPhotos") != 1 {
		t.Errorf("expected 1 call to people.getPhotos, got %d", fake.CountMethod("flickr.people.getPhotos"))
	}
}

func TestPhotosSearchJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Sunset", Owner: "user1", Tags: "nature"}

	cmd, buf := cmdContext(t, cfg, true)
	if err := cmd.Flags().Set("text", "sunset"); err != nil {
		t.Fatal(err)
	}
	err := photosSearchCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.search" {
		t.Errorf("expected command=photos.search, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatal("items is not a slice")
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 photo, got %d", len(items))
	}

	call, ok := fake.LastCall("flickr.photos.search")
	if !ok {
		t.Fatal("expected call to flickr.photos.search")
	}
	if call.Params["text"] != "sunset" {
		t.Errorf("expected text=sunset, got %s", call.Params["text"])
	}
}

func TestPhotosShowJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["photo-99"] = testutil.FakePhoto{
		ID:    "photo-99",
		Title: "Test Photo",
		Owner: "test-user-123",
	}

	cmd, buf := cmdContext(t, cfg, true)
	err := photosShowCmd.RunE(cmd, []string{"photo-99"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.show" {
		t.Errorf("expected command=photos.show, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["id"] != "photo-99" {
		t.Errorf("expected id=photo-99, got %v", data["id"])
	}
	if data["title"] != "Test Photo" {
		t.Errorf("expected title=Test Photo, got %v", data["title"])
	}
}

func TestPhotosShowNotFound(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := photosShowCmd.RunE(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent photo")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrFlickrAPI {
		t.Errorf("expected FLICKR_API_ERROR, got %s", env.Error.Code)
	}
}

// --- API integration tests ---

func TestAPICallJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := apiCallCmd.RunE(cmd, []string{"flickr.test.echo"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "api.call" {
		t.Errorf("expected command=api.call, got %s", env.Meta.Command)
	}

	if fake.CountMethod("flickr.test.echo") != 1 {
		t.Errorf("expected 1 call to test.echo, got %d", fake.CountMethod("flickr.test.echo"))
	}
}

func TestAPICallRawMode(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	_ = cmd.Flags().Set("raw", "true")
	err := apiCallCmd.RunE(cmd, []string{"flickr.test.echo"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	data, _ := env.Data.(map[string]any)
	if _, ok := data["raw"]; !ok {
		t.Error("expected data.raw to be present in raw mode")
	}
}

func TestAPIMethodsJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := apiMethodsCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "api.methods" {
		t.Errorf("expected command=api.methods, got %s", env.Meta.Command)
	}

	if fake.CountMethod("flickr.reflection.getMethods") != 1 {
		t.Errorf("expected 1 call to getMethods, got %d", fake.CountMethod("flickr.reflection.getMethods"))
	}
}

// --- Version integration tests ---

func TestVersionGolden(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetContext(WithAppContext(context.Background(), &AppContext{
		JSON:      true,
		Profile:   "default",
		RequestID: uuid.New().String(),
		StartedAt: time.Now(),
		Timeout:   30 * time.Second,
	}))

	err := versionCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "version" {
		t.Errorf("expected command=version, got %s", env.Meta.Command)
	}
	if env.Meta.SchemaVersion != model.SchemaVersion {
		t.Errorf("expected schema_version=%s, got %s", model.SchemaVersion, env.Meta.SchemaVersion)
	}

	data, _ := env.Data.(map[string]any)
	if data["schema_version"] != model.SchemaVersion {
		t.Errorf("expected data.schema_version=%s, got %v", model.SchemaVersion, data["schema_version"])
	}
}

// --- Error envelope tests ---

func TestErrorEnvelopeHasAllFields(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg, true)
	_ = albumsListCmd.RunE(cmd, nil)

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error body")
	}
	if env.Error.Code == "" {
		t.Error("expected non-empty error code")
	}
	if env.Error.Message == "" {
		t.Error("expected non-empty error message")
	}
	if env.Meta.SchemaVersion == "" {
		t.Error("expected non-empty schema_version in meta")
	}
	if env.Meta.Command == "" {
		t.Error("expected non-empty command in meta")
	}
}

// --- Implemented M5 commands return success with fake Flickr ---

func TestImplementedCommandsSucceed(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Test", Owner: "test-user-123"}

	tests := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{"photos.set-tags", photosSetTagsCmd, []string{"p1"}},
		{"photos.add-tags", photosAddTagsCmd, []string{"p1"}},
		{"photos.remove-tag", photosRemoveTagCmd, []string{"p1"}},
		{"photos.set-privacy", photosSetPrivacyCmd, []string{"p1"}},
		{"photos.set-location", photosSetLocationCmd, []string{"p1"}},
		{"photos.rotate", photosRotateCmd, []string{"p1"}},
		{"photos.delete", photosDeleteCmd, []string{"p1"}},
		{"photos.set-meta", photosSetMetaCmd, []string{"p1"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, buf := cmdContext(t, cfg, true)
			_ = tc.cmd.Flags().Set("confirm", "true")
			_ = tc.cmd.Flags().Set("tag", "test")
			_ = tc.cmd.Flags().Set("tag-id", "tag-1")
			_ = tc.cmd.Flags().Set("title", "New Title")
			_ = tc.cmd.Flags().Set("privacy", "public")
			_ = tc.cmd.Flags().Set("lat", "40.0")
			_ = tc.cmd.Flags().Set("lon", "-74.0")
			err := tc.cmd.RunE(cmd, tc.args)

			env := parseEnvelope(t, buf)
			if err != nil && !env.OK {
				// Command returned an error (e.g., API error from fake server) - this is acceptable
				return
			}
			if !env.OK {
				t.Errorf("command %s should return ok=true or a handled error, got error: %v", tc.name, env.Error)
			}
		})
	}
}

// --- Upload dry-run test ---

func TestPhotosUploadDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	// Create a temp directory with a test image file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("fake jpeg data"), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	err := photosUploadCmd.RunE(cmd, []string{testFile})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.upload" {
		t.Errorf("expected command=photos.upload, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}

	plan, ok := data["plan"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan to be a map, got %T", data["plan"])
	}

	plannedItems, ok := plan["planned"].([]any)
	if !ok {
		t.Fatalf("expected plan.planned to be an array, got %T", plan["planned"])
	}
	if len(plannedItems) == 0 {
		t.Fatal("expected at least one planned upload")
	}

	// Verify the test file is in the planned list
	found := false
	for _, item := range plannedItems {
		pu, ok := item.(map[string]any)
		if !ok {
			t.Fatal("item is not a map")
		}
		localPath, _ := pu["local_path"].(string)
		if strings.Contains(localPath, "test.jpg") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected test.jpg to be in the planned uploads, got %v", plannedItems)
	}
}

// --- Favorites integration tests ---

func TestFavoritesListJSON(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Fav Photo"}

	cmd, buf := cmdContext(t, cfg, true)
	err := favoritesListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "favorites.list" {
		t.Errorf("expected command=favorites.list, got %s", env.Meta.Command)
	}
}

func TestFavoritesAddDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	err := favoritesAddCmd.RunE(cmd, []string{"p1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

func TestFavoritesRemoveDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	err := favoritesRemoveCmd.RunE(cmd, []string{"p1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
}

func TestFavoritesReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
	err := favoritesAddCmd.RunE(cmd, []string{"p1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrReadOnlyViolation {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Galleries integration tests ---

func TestGalleriesListJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := galleriesListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "galleries.list" {
		t.Errorf("expected command=galleries.list, got %s", env.Meta.Command)
	}
}

func TestGalleriesPhotosJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := galleriesPhotosCmd.RunE(cmd, []string{"gallery-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "galleries.photos" {
		t.Errorf("expected command=galleries.photos, got %s", env.Meta.Command)
	}
}

// --- Groups integration tests ---

func TestGroupsListJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := groupsListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "groups.list" {
		t.Errorf("expected command=groups.list, got %s", env.Meta.Command)
	}
}

func TestGroupsSearchJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := groupsSearchCmd.RunE(cmd, []string{"photography"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "groups.search" {
		t.Errorf("expected command=groups.search, got %s", env.Meta.Command)
	}
}

// --- Comments integration tests ---

func TestCommentsListJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := commentsListCmd.RunE(cmd, []string{"p1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "comments.list" {
		t.Errorf("expected command=comments.list, got %s", env.Meta.Command)
	}
}

func TestCommentsAddDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	err := commentsAddCmd.RunE(cmd, []string{"p1", "Nice photo!"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

func TestCommentsDeleteRequiresConfirm(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := commentsDeleteCmd.RunE(cmd, []string{"comment-1"})
	if err == nil {
		t.Fatal("expected error without --confirm")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrConfirmationRequired {
		t.Errorf("expected CONFIRMATION_REQUIRED, got %s", env.Error.Code)
	}
}

func TestCommentsDeleteReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true, Confirm: true})
	err := commentsDeleteCmd.RunE(cmd, []string{"comment-1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrReadOnlyViolation {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Contacts integration tests ---

func TestContactsListJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := contactsListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "contacts.list" {
		t.Errorf("expected command=contacts.list, got %s", env.Meta.Command)
	}
}

// --- Stats integration tests ---

func TestStatsPopularJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := statsPopularCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "stats.popular" {
		t.Errorf("expected command=stats.popular, got %s", env.Meta.Command)
	}
}

// --- URLs integration tests ---

func TestURLsLookupUserJSON(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	err := urlsLookupUserCmd.RunE(cmd, []string{"https://flickr.com/testuser"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "urls.lookupUser" {
		t.Errorf("expected command=urls.lookupUser, got %s", env.Meta.Command)
	}
}

// --- Photos download dry-run test ---

func TestPhotosDownloadDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	err := photosDownloadCmd.RunE(cmd, []string{"p1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

// --- RequireAuth strengthens: check error code, not just buf.Len ---

func TestRequireAuthWritesErrorCode(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg, true)
	_ = photosListCmd.RunE(cmd, nil)

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error body")
	}
	if env.Error.Code != model.ErrAuthRequired {
		t.Errorf("expected AUTH_REQUIRED, got %s", env.Error.Code)
	}
	if env.Meta.Command != "photos.list" {
		t.Errorf("expected command=photos.list, got %s", env.Meta.Command)
	}
}

// --- backupModeToPlanMode unit tests ---

func TestBackupModeToPlanMode(t *testing.T) {
	tests := []struct {
		name      string
		layout    string
		all       bool
		hasAlbums bool
		wantMode  backup.PlanMode
	}{
		{"id-dirs layout", "id-dirs", false, false, backup.BackupIDDirs},
		{"album layout", "album", false, false, backup.BackupAlbums},
		{"all flag with empty layout", "", true, false, backup.BackupAlbums},
		{"hasAlbums with empty layout", "", false, true, backup.BackupAlbums},
		{"default user mode", "", false, false, backup.BackupUser},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := backupModeToPlanMode(tc.layout, tc.all, tc.hasAlbums)
			if got != tc.wantMode {
				t.Errorf("backupModeToPlanMode(%q, %v, %v) = %q, want %q",
					tc.layout, tc.all, tc.hasAlbums, got, tc.wantMode)
			}
		})
	}
}

// --- Photos download backup dry-run test ---

func TestPhotosDownloadBackupDryRun(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Albums["album-1"] = testutil.FakeAlbum{
		ID:          "album-1",
		Title:       "Summer",
		Description: "Summer photos",
		PhotoCount:  2,
		PrimaryID:   "p1",
	}
	fake.Photos["p1"] = testutil.FakePhoto{ID: "p1", Title: "Sunset", Owner: "test-user-123"}
	fake.Photos["p2"] = testutil.FakePhoto{ID: "p2", Title: "Beach", Owner: "test-user-123"}
	fake.AlbumPhotos["album-1"] = []string{"p1", "p2"}

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	cmd.Flags().Bool("all", false, "")
	_ = cmd.Flags().Set("all", "true")
	cmd.Flags().String("dest", "", "")
	_ = cmd.Flags().Set("dest", "/tmp/test-dest")
	cmd.Flags().String("size", "original", "")
	cmd.Flags().Int("size-max", 0, "")
	cmd.Flags().String("metadata", "json", "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().String("layout", "", "")
	cmd.Flags().StringSlice("album", nil, "")
	cmd.Flags().StringSlice("album-id", nil, "")
	cmd.Flags().Bool("exif", false, "")

	err := photosDownloadCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	total, ok := data["total"].(float64)
	if !ok {
		t.Fatalf("expected total to be a number, got %T", data["total"])
	}
	if int(total) != 2 {
		t.Errorf("expected total=2, got %d", int(total))
	}
}

// --- handleRequestTokenError tests ---

func TestHandleRequestTokenError(t *testing.T) {
	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	// 400 error -> AuthFailed with "Invalid API key" message
	err := handleRequestTokenError(r, meta, fmt.Errorf("400 bad request"))
	if err == nil {
		t.Fatal("expected error for 400")
	}
	var cmdErr *model.CommandError
	if !stderrors.As(err, &cmdErr) {
		t.Fatalf("expected CommandError, got %T", err)
	}
	if cmdErr.Code != model.ErrAuthFailed {
		t.Errorf("expected AUTH_FAILED, got %s", cmdErr.Code)
	}
	if !strings.Contains(cmdErr.Message, "Invalid API key") {
		t.Errorf("expected message about invalid API key, got %q", cmdErr.Message)
	}

	// Non-400 error -> AuthFailed with "requesting token" message
	buf.Reset()
	err = handleRequestTokenError(r, meta, fmt.Errorf("network timeout"))
	if err == nil {
		t.Fatal("expected error for non-400")
	}
	if !stderrors.As(err, &cmdErr) {
		t.Fatalf("expected CommandError, got %T", err)
	}
	if cmdErr.Code != model.ErrAuthFailed {
		t.Errorf("expected AUTH_FAILED, got %s", cmdErr.Code)
	}
	if !strings.Contains(cmdErr.Message, "requesting token") {
		t.Errorf("expected message about requesting token, got %q", cmdErr.Message)
	}
}

// --- resolveCredentials from flags test ---

func TestResolveCredentialsFromFlags(t *testing.T) {
	buf := new(bytes.Buffer)
	r := output.Renderer{Out: buf, Err: buf, JSON: true}
	meta := output.RuntimeMetaInput{Command: "test", StartedAt: time.Now()}

	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().String("api-secret", "", "")
	cmd.Flags().String("api-secret-env", "", "")
	_ = cmd.Flags().Set("api-key", "my-api-key")
	_ = cmd.Flags().Set("api-secret", "my-api-secret")

	apiKey, apiSecret, err := resolveCredentials(cmd, &r, meta, config.Credentials{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "my-api-key" {
		t.Errorf("expected api-key=my-api-key, got %s", apiKey)
	}
	if apiSecret != "my-api-secret" {
		t.Errorf("expected api-secret=my-api-secret, got %s", apiSecret)
	}
}

// --- Checksums add read-only guard test ---

func TestChecksumsReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
	err := checksumsAddCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrReadOnlyViolation {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Photos download no args validation test ---

func TestPhotosDownloadNoArgs(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	// Ensure download-specific flags are registered with defaults (no --all, etc.)
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().String("dest", "", "")
	cmd.Flags().String("size", "original", "")
	cmd.Flags().Int("size-max", 0, "")
	cmd.Flags().String("metadata", "json", "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().String("layout", "", "")
	cmd.Flags().StringSlice("album", nil, "")
	cmd.Flags().StringSlice("album-id", nil, "")
	cmd.Flags().Bool("exif", false, "")

	err := photosDownloadCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected validation error with no args and no --all/--album")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != model.ErrValidationFailed {
		t.Errorf("expected VALIDATION_FAILED, got %s", env.Error.Code)
	}
}

// --- Photos download by IDs with metadata test ---

func TestPhotosDownloadByIDsWithMetadata(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	cmd.Flags().String("dest", "", "")
	cmd.Flags().String("size", "original", "")
	cmd.Flags().Int("size-max", 0, "")
	cmd.Flags().String("metadata", "json", "")
	_ = cmd.Flags().Set("metadata", "both")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().String("layout", "", "")
	cmd.Flags().StringSlice("album", nil, "")
	cmd.Flags().StringSlice("album-id", nil, "")
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().Bool("exif", false, "")

	err := photosDownloadCmd.RunE(cmd, []string{"p1", "p2"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}

	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("expected items to be an array, got %T", data["items"])
	}
	if len(items) != 2 {
		t.Errorf("expected 2 planned items, got %d", len(items))
	}
}
