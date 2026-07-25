package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/testutil"
)

func TestPiwigoHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"piwigo", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestPiwigoImportMissingFlags(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
	// Register piwigo flags so RunE can read them with their zero values
	cmd.Flags().String("url", "", "")
	cmd.Flags().String("user", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("album-prefix", "", "")
	cmd.Flags().String("import-album", "", "")
	cmd.Flags().String("dedupe", "", "")
	cmd.Flags().Int("limit", 0, "")

	// Call without setting --url, --user, or --password
	err := piwigoImportCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing required flags")
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

func registerPiwigoFlags(cmd *cobra.Command, url string) {
	cmd.Flags().String("url", "", "")
	cmd.Flags().String("user", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("album-prefix", "", "")
	cmd.Flags().String("import-album", "Imported from Piwigo", "")
	cmd.Flags().String("dedupe", "checksum", "")
	cmd.Flags().Int("limit", 0, "")
	_ = cmd.Flags().Set("url", url)
	_ = cmd.Flags().Set("user", "admin")
	_ = cmd.Flags().Set("password", "secret")
}

func TestPiwigoImportDryRunPlan(t *testing.T) {
	fake, cfg := setupFakeCLI(t)

	pw := testutil.NewFakePiwigo(t)
	pw.Categories = []testutil.FakePiwigoCategory{
		{ID: "cat1", Name: "Vacation", NbImages: 2},
		{ID: "cat2", Name: "Nature", NbImages: 1},
	}
	pw.Images = map[string][]testutil.FakePiwigoImage{
		"cat1": {
			{ID: "img1", File: "a.jpg", Name: "A", MD5Sum: "aaa", CategoryIDs: []string{"cat1"}},
			{ID: "img2", File: "b.jpg", Name: "B", MD5Sum: "bbb", CategoryIDs: []string{"cat1"}},
		},
		"cat2": {
			{ID: "img3", File: "c.jpg", Name: "C", MD5Sum: "ccc", CategoryIDs: []string{"cat2"}},
		},
	}

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	registerPiwigoFlags(cmd, pw.Server.URL)

	if err := piwigoImportCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "piwigo.import" {
		t.Errorf("expected command=piwigo.import, got %s", env.Meta.Command)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", env.Data)
	}
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	if got := data["planned_photos"]; got != float64(3) {
		t.Errorf("planned_photos = %v, want 3", got)
	}
	if got := data["planned_albums"]; got != float64(3) {
		t.Errorf("planned_albums = %v, want 3", got)
	}
	if got := data["skipped"]; got != float64(0) {
		t.Errorf("skipped = %v, want 0", got)
	}

	if fake.CountMethod("upload") != 0 {
		t.Errorf("expected 0 Flickr uploads in dry-run, got %d", fake.CountMethod("upload"))
	}
	if len(fake.Calls) != 0 {
		t.Errorf("expected no Flickr calls in dry-run, got %+v", fake.Calls)
	}
}

func TestPiwigoImportDryRunSkipsExisting(t *testing.T) {
	fake, cfg := setupFakeCLI(t)

	pw := testutil.NewFakePiwigo(t)
	pw.Categories = []testutil.FakePiwigoCategory{{ID: "cat1", Name: "Vacation", NbImages: 2}}
	pw.Images = map[string][]testutil.FakePiwigoImage{
		"cat1": {
			{ID: "img1", File: "dup.jpg", Name: "Dup", MD5Sum: "dup", CategoryIDs: []string{"cat1"}},
			{ID: "img2", File: "new.jpg", Name: "New", MD5Sum: "new", CategoryIDs: []string{"cat1"}},
		},
	}
	pw.ExistingMD5 = map[string]bool{"dup": true}

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
	registerPiwigoFlags(cmd, pw.Server.URL)

	if err := piwigoImportCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", env.Data)
	}
	if got := data["planned_photos"]; got != float64(1) {
		t.Errorf("planned_photos = %v, want 1", got)
	}
	if got := data["skipped"]; got != float64(1) {
		t.Errorf("skipped = %v, want 1", got)
	}
	if got := data["planned_albums"]; got != float64(2) {
		t.Errorf("planned_albums = %v, want 2", got)
	}
	if fake.CountMethod("upload") != 0 {
		t.Errorf("expected 0 Flickr uploads in dry-run, got %d", fake.CountMethod("upload"))
	}
}
