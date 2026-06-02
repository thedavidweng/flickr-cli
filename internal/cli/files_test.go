package cli

import (
	"bytes"
	"testing"
)

func TestFilesHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"files", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestFilesListDryRun(t *testing.T) {
	fake, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	cmd.Flags().StringSlice("album", nil, "")
	cmd.Flags().StringSlice("album-id", nil, "")
	cmd.Flags().Set("album-id", "album-1")
	err := filesListCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "files.list" {
		t.Errorf("expected command=files.list, got %s", env.Meta.Command)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be a map, got %T", env.Data)
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("expected data.items to be an array, got %T", data["items"])
	}
	// Fake server returns empty photos for getAlbumPhotos
	if len(items) != 0 {
		t.Errorf("expected 0 photo IDs (fake album has no photos), got %d", len(items))
	}

	// Verify that getPhotos was called
	if fake.CountMethod("flickr.photosets.getPhotos") != 1 {
		t.Errorf("expected 1 call to getPhotos, got %d", fake.CountMethod("flickr.photosets.getPhotos"))
	}
}
