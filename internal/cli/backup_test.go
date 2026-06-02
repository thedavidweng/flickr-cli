package cli

import (
	"bytes"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/testutil"
	"github.com/spf13/cobra"
)

func TestBackupHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backup", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

// setupBackupAlbumsCmd creates a test cobra.Command with all backupAlbumsCmd flags registered.
func setupBackupAlbumsCmd(t *testing.T, cfgPath string, jsonMode bool, appOverrides ...*AppContext) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	cmd, buf := cmdContext(t, cfgPath, jsonMode, appOverrides...)
	cmd.Flags().String("dest", "./flickr-backup", "")
	cmd.Flags().StringSlice("album", nil, "")
	cmd.Flags().StringSlice("album-id", nil, "")
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().String("size", "original", "")
	cmd.Flags().String("metadata", "json", "")
	cmd.Flags().String("template", "archive", "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("include-comments", false, "")
	cmd.Flags().Bool("include-geo", false, "")
	cmd.Flags().Bool("include-pools", false, "")
	cmd.Flags().Bool("include-albums", false, "")
	cmd.Flags().Bool("resume", false, "")
	return cmd, buf
}

func TestBackupAlbumsDryRun(t *testing.T) {
	fake, cfg := setupFakeCLI(t)
	fake.Albums["album-1"] = testutil.FakeAlbum{
		ID:         "album-1",
		Title:      "Summer Vacation",
		PhotoCount: 10,
		PrimaryID:  "photo-1",
	}

	cmd, buf := setupBackupAlbumsCmd(t, cfg, true, &AppContext{DryRun: true})
	cmd.Flags().Set("all", "true")
	err := backupAlbumsCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "backup.albums" {
		t.Errorf("expected command=backup.albums, got %s", env.Meta.Command)
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
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0].(map[string]any)
	if item["photo_id"] != "album-1" {
		t.Errorf("expected photo_id=album-1, got %v", item["photo_id"])
	}
	if item["title"] != "Summer Vacation" {
		t.Errorf("expected title=Summer Vacation, got %v", item["title"])
	}

	// Verify the plan was built (getList called) but no download happened
	if fake.CountMethod("flickr.photosets.getList") != 1 {
		t.Errorf("expected 1 call to getList, got %d", fake.CountMethod("flickr.photosets.getList"))
	}
	if fake.CountMethod("flickr.photos.getSizes") != 0 {
		t.Errorf("expected 0 calls to getSizes in dry-run, got %d", fake.CountMethod("flickr.photos.getSizes"))
	}
}

func TestBackupAlbumsRequiresSelection(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := setupBackupAlbumsCmd(t, cfg, true)
	err := backupAlbumsCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error without --all, --album, or --album-id")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error == nil {
		t.Fatal("expected error body")
	}
	if env.Error.Code != model.ErrValidationFailed {
		t.Errorf("expected VALIDATION_FAILED, got %s", env.Error.Code)
	}
}
