package cli

import (
	"bytes"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func TestChecksumsHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"checksums", "--help"})
	rootCmd.Execute()

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
