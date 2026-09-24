package cli

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// --- Photos delete dry-run ---

func TestPhotosDeleteDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	err := photosDeleteCmd.RunE(cmd, []string{"photo-a", "photo-b"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.delete" {
		t.Errorf("expected command=photos.delete, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}

	ids, ok := data["photo_ids"].([]any)
	if !ok {
		t.Fatalf("expected photo_ids to be a slice, got %T", data["photo_ids"])
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 photo ids, got %d", len(ids))
	}
}

// --- Photos delete requires confirm ---

func TestPhotosDeleteRequiresConfirm(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg)
	err := photosDeleteCmd.RunE(cmd, []string{"photo-1"})
	if err == nil {
		t.Fatal("expected error without --confirm")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "CONFIRMATION_REQUIRED" {
		t.Errorf("expected CONFIRMATION_REQUIRED, got %s", env.Error.Code)
	}
}

// --- Photos delete read-only ---

func TestPhotosDeleteReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{ReadOnly: true, Confirm: true})
	err := photosDeleteCmd.RunE(cmd, []string{"photo-1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "READ_ONLY_VIOLATION" {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Photos rotate dry-run ---

func TestPhotosRotateDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	cmd.Flags().Int("degrees", 90, "")
	err := photosRotateCmd.RunE(cmd, []string{"photo-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.rotate" {
		t.Errorf("expected command=photos.rotate, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	if data["degrees"] != float64(90) {
		t.Errorf("expected degrees=90, got %v", data["degrees"])
	}
}

// --- Photos rotate read-only ---

func TestPhotosRotateReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{ReadOnly: true})
	err := photosRotateCmd.RunE(cmd, []string{"photo-1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "READ_ONLY_VIOLATION" {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Photos set-privacy dry-run ---

func TestPhotosPrivacyDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	err := photosSetPrivacyCmd.RunE(cmd, []string{"photo-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.set-privacy" {
		t.Errorf("expected command=photos.set-privacy, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

// --- Photos set-privacy read-only ---

func TestPhotosPrivacyReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{ReadOnly: true})
	err := photosSetPrivacyCmd.RunE(cmd, []string{"photo-1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "READ_ONLY_VIOLATION" {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Photos set-meta dry-run ---

func TestPhotosMetaDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	if err := cmd.Flags().Set("title", "New Title"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("description", "New Desc"); err != nil {
		t.Fatal(err)
	}
	err := photosSetMetaCmd.RunE(cmd, []string{"photo-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.set-meta" {
		t.Errorf("expected command=photos.set-meta, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	if data["title"] != "New Title" {
		t.Errorf("expected title=New Title, got %v", data["title"])
	}
	if data["description"] != "New Desc" {
		t.Errorf("expected description=New Desc, got %v", data["description"])
	}
}

// --- Photos set-meta read-only ---

func TestPhotosMetaReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{ReadOnly: true})
	if err := cmd.Flags().Set("title", "T"); err != nil {
		t.Fatal(err)
	}
	err := photosSetMetaCmd.RunE(cmd, []string{"photo-1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "READ_ONLY_VIOLATION" {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Photos set-tags dry-run ---

func TestPhotosSetTagsDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	err := photosSetTagsCmd.RunE(cmd, []string{"photo-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.set-tags" {
		t.Errorf("expected command=photos.set-tags, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

// --- Photos add-tags dry-run ---

func TestPhotosAddTagsDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	err := photosAddTagsCmd.RunE(cmd, []string{"photo-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.add-tags" {
		t.Errorf("expected command=photos.add-tags, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

// --- Photos remove-tag dry-run ---

func TestPhotosRemoveTagDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	cmd.Flags().String("tag-id", "", "")
	if err := cmd.Flags().Set("tag-id", "tag-1"); err != nil {
		t.Fatal(err)
	}
	err := photosRemoveTagCmd.RunE(cmd, []string{"photo-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "photos.remove-tag" {
		t.Errorf("expected command=photos.remove-tag, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
	if data["tag_id"] != "tag-1" {
		t.Errorf("expected tag_id=tag-1, got %v", data["tag_id"])
	}
}

// --- Comments delete dry-run ---

func TestCommentsDeleteDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	err := commentsDeleteCmd.RunE(cmd, []string{"comment-1"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
	}
	if env.Meta.Command != "comments.delete" {
		t.Errorf("expected command=comments.delete, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

// --- Help output tests ---
// These verify the command's Short description and Use line are present in help.

// --- Photos parent command help ---

// --- Galleries parent command help ---

// --- Groups parent command help ---

// --- Comments parent command help ---

// --- Favorites parent command help ---

// --- Photos rotate validation error ---

func TestPhotosRotateInvalidDegrees(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg)
	cmd.Flags().Int("degrees", 90, "")
	if err := cmd.Flags().Set("degrees", "45"); err != nil {
		t.Fatal(err)
	}
	err := photosRotateCmd.RunE(cmd, []string{"photo-1"})
	if err == nil {
		t.Fatal("expected error for invalid degrees")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "VALIDATION_FAILED" {
		t.Errorf("expected VALIDATION_FAILED, got %s", env.Error.Code)
	}
}

// --- Stats popular help ---

// --- URLs lookup-user requires auth ---

func TestURLsLookupUserAuthRequired(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg)
	err := urlsLookupUserCmd.RunE(cmd, []string{"https://flickr.com/testuser"})
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "AUTH_REQUIRED" {
		t.Errorf("expected AUTH_REQUIRED, got %s", env.Error.Code)
	}
}

// --- Gallery photos auth required ---

func TestGalleriesPhotosAuthRequired(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg)
	err := galleriesPhotosCmd.RunE(cmd, []string{"gallery-1"})
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "AUTH_REQUIRED" {
		t.Errorf("expected AUTH_REQUIRED, got %s", env.Error.Code)
	}
}

// --- Groups search auth required ---

func TestGroupsSearchAuthRequired(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg)
	err := groupsSearchCmd.RunE(cmd, []string{"photography"})
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "AUTH_REQUIRED" {
		t.Errorf("expected AUTH_REQUIRED, got %s", env.Error.Code)
	}
}

// --- Stats popular auth required ---

func TestStatsPopularAuthRequired(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg)
	err := statsPopularCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "AUTH_REQUIRED" {
		t.Errorf("expected AUTH_REQUIRED, got %s", env.Error.Code)
	}
}

// --- Favorites add blocked by read-only ---

func TestFavoritesAddReadOnly(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{ReadOnly: true})
	err := favoritesAddCmd.RunE(cmd, []string{"p1"})
	if err == nil {
		t.Fatal("expected error with --read-only")
	}

	env := parseEnvelope(t, buf)
	if env.OK {
		t.Fatal("expected ok=false")
	}
	if env.Error.Code != "READ_ONLY_VIOLATION" {
		t.Errorf("expected READ_ONLY_VIOLATION, got %s", env.Error.Code)
	}
}

// --- Comments add dry-run ---

func TestCommentsAddDryRunMeta(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
	err := commentsAddCmd.RunE(cmd, []string{"photo-1", "Great shot!"})
	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	env := parseEnvelope(t, buf)
	if !env.OK {
		t.Fatalf("expected ok=true, got error: %v", env.Error)
	}
	if env.Meta.Command != "comments.add" {
		t.Errorf("expected command=comments.add, got %s", env.Meta.Command)
	}

	data, _ := env.Data.(map[string]any)
	if data["planned"] != true {
		t.Errorf("expected planned=true, got %v", data["planned"])
	}
}

// --- Root command has all expected subcommands ---

func TestRootCommandSubcommands(t *testing.T) {
	root := rootCmd
	expected := []string{
		"version", "auth", "albums", "photos", "favorites",
		"galleries", "groups", "comments", "contacts",
		"stats", "urls", "api", "cache", "checksums",
		"piwigo", "doctor", "completion",
	}
	names := make(map[string]bool)
	for _, sub := range root.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range expected {
		if !names[want] {
			t.Errorf("root command missing subcommand %q", want)
		}
	}
}

// --- AppContext via WithAppContext round-trip ---

func TestAppContextRoundTrip(t *testing.T) {
	app := &AppContext{
		ConfigFile: "/tmp/test.yaml",
		Profile:    "myprofile",
		JSON:       true,
		ReadOnly:   true,
		DryRun:     false,
		Confirm:    true,
		RequestID:  "req-123",
		Timeout:    10 * time.Second,
		Retries:    5,
	}
	ctx := WithAppContext(context.Background(), app)
	got := GetAppContext(ctx)
	if got != app {
		t.Fatal("GetAppContext should return the same pointer")
	}
	if got.ConfigFile != "/tmp/test.yaml" {
		t.Errorf("expected ConfigFile=/tmp/test.yaml, got %s", got.ConfigFile)
	}
	if got.Profile != "myprofile" {
		t.Errorf("expected Profile=myprofile, got %s", got.Profile)
	}
	if !got.JSON {
		t.Error("expected JSON=true")
	}
	if !got.ReadOnly {
		t.Error("expected ReadOnly=true")
	}
	if !got.Confirm {
		t.Error("expected Confirm=true")
	}
	if got.Retries != 5 {
		t.Errorf("expected Retries=5, got %d", got.Retries)
	}
}

// --- GetAppContext with nil returns nil ---

func TestGetAppContextNil(t *testing.T) {
	got := GetAppContext(context.Background())
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// --- All mutation commands produce dry-run envelope ---

func TestAllMutationCommandsDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	tests := []struct {
		name    string
		cmd     *cobra.Command
		args    []string
		command string
		flags   map[string]string
	}{
		{"photos.delete", photosDeleteCmd, []string{"p1"}, "photos.delete", nil},
		{"photos.rotate", photosRotateCmd, []string{"p1"}, "photos.rotate", nil},
		{"photos.set-privacy", photosSetPrivacyCmd, []string{"p1"}, "photos.set-privacy", nil},
		{"photos.set-meta", photosSetMetaCmd, []string{"p1"}, "photos.set-meta", map[string]string{"title": "T"}},
		{"photos.set-tags", photosSetTagsCmd, []string{"p1"}, "photos.set-tags", nil},
		{"photos.add-tags", photosAddTagsCmd, []string{"p1"}, "photos.add-tags", nil},
		{"photos.remove-tag", photosRemoveTagCmd, []string{"p1"}, "photos.remove-tag", map[string]string{"tag-id": "t1"}},
		{"comments.add", commentsAddCmd, []string{"p1", "text"}, "comments.add", nil},
		{"comments.delete", commentsDeleteCmd, []string{"c1"}, "comments.delete", nil},
		{"favorites.add", favoritesAddCmd, []string{"p1"}, "favorites.add", nil},
		{"favorites.remove", favoritesRemoveCmd, []string{"p1"}, "favorites.remove", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, buf := cmdContext(t, cfg, &AppContext{DryRun: true})
			// Register flags that specific commands read but cmdContext doesn't provide
			cmd.Flags().Int("degrees", 90, "")
			cmd.Flags().String("tag-id", "", "")
			cmd.Flags().String("privacy", "", "")
			cmd.Flags().String("hidden", "", "")
			cmd.Flags().StringSlice("tag", nil, "")
			cmd.Flags().String("tags", "", "")
			cmd.Flags().Float64("lat", 0, "")
			cmd.Flags().Float64("lon", 0, "")
			cmd.Flags().Int("accuracy", 0, "")
			cmd.Flags().Int("context", 0, "")
			for k, v := range tc.flags {
				if err := cmd.Flags().Set(k, v); err != nil {
					t.Fatalf("setting flag %s: %v", k, err)
				}
			}
			err := tc.cmd.RunE(cmd, tc.args)
			if err != nil {
				t.Fatalf("RunE returned error: %v", err)
			}

			env := parseEnvelope(t, buf)
			if !env.OK {
				t.Fatalf("expected ok=true for dry-run, got error: %v", env.Error)
			}
			if env.Meta.Command != tc.command {
				t.Errorf("expected command=%s, got %s", tc.command, env.Meta.Command)
			}

			data, _ := env.Data.(map[string]any)
			if data["planned"] != true {
				t.Errorf("expected planned=true, got %v", data["planned"])
			}
		})
	}
}
