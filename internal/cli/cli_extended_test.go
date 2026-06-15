package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// cmdHelp returns the help text for a command by setting up a buffer and
// calling cmd.Help().  The command must already have Use/Short set.
func cmdHelp(t *testing.T, cmd *cobra.Command) string {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Help(); err != nil {
		t.Fatalf("Help() returned error: %v", err)
	}
	return buf.String()
}

// subcommandHelp returns the help text for a specific subcommand of a parent.
func subcommandHelp(t *testing.T, parent *cobra.Command, subName string) string {
	t.Helper()
	for _, sub := range parent.Commands() {
		if sub.Name() == subName {
			return cmdHelp(t, sub)
		}
	}
	t.Fatalf("subcommand %q not found on %s", subName, parent.Name())
	return ""
}

// --- Photos delete dry-run ---

func TestPhotosDeleteDryRun(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true)
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true, Confirm: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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

func TestPhotosRotateHelp(t *testing.T) {
	help := cmdHelp(t, photosRotateCmd)
	if !strings.Contains(help, "rotate") {
		t.Error("help should contain 'rotate'")
	}
	if !strings.Contains(help, "[photo-id]") {
		t.Error("help should contain '[photo-id]' usage")
	}
	if !strings.Contains(help, "degrees") {
		t.Error("help should mention --degrees flag")
	}
}

func TestPhotosPrivacyHelp(t *testing.T) {
	help := cmdHelp(t, photosSetPrivacyCmd)
	if !strings.Contains(help, "set-privacy") {
		t.Error("help should contain 'set-privacy'")
	}
	if !strings.Contains(help, "[photo-id]") {
		t.Error("help should contain '[photo-id]' usage")
	}
	if !strings.Contains(help, "privacy") {
		t.Error("help should mention --privacy flag")
	}
}

func TestPhotosMetaHelp(t *testing.T) {
	help := cmdHelp(t, photosSetMetaCmd)
	if !strings.Contains(help, "set-meta") {
		t.Error("help should contain 'set-meta'")
	}
	if !strings.Contains(help, "[photo-id]") {
		t.Error("help should contain '[photo-id]' usage")
	}
	if !strings.Contains(help, "title") {
		t.Error("help should mention --title flag")
	}
}

func TestGalleriesListHelp(t *testing.T) {
	help := cmdHelp(t, galleriesListCmd)
	if !strings.Contains(help, "list") {
		t.Error("help should contain 'list'")
	}
	if !strings.Contains(strings.ToLower(help), "galler") {
		t.Error("help should mention galleries")
	}
}

func TestGroupsListHelp(t *testing.T) {
	help := cmdHelp(t, groupsListCmd)
	if !strings.Contains(help, "list") {
		t.Error("help should contain 'list'")
	}
	if !strings.Contains(help, "group") && !strings.Contains(help, "Group") {
		t.Error("help should mention groups")
	}
}

func TestContactsListHelp(t *testing.T) {
	help := cmdHelp(t, contactsListCmd)
	if !strings.Contains(help, "list") {
		t.Error("help should contain 'list'")
	}
	if !strings.Contains(help, "contact") && !strings.Contains(help, "Contact") {
		t.Error("help should mention contacts")
	}
}

func TestCommentsListHelp(t *testing.T) {
	help := cmdHelp(t, commentsListCmd)
	if !strings.Contains(help, "list") {
		t.Error("help should contain 'list'")
	}
	if !strings.Contains(help, "[photo-id]") {
		t.Error("help should contain '[photo-id]' usage")
	}
	if !strings.Contains(help, "comment") && !strings.Contains(help, "Comment") {
		t.Error("help should mention comments")
	}
}

func TestFavoritesListHelp(t *testing.T) {
	help := cmdHelp(t, favoritesListCmd)
	if !strings.Contains(help, "list") {
		t.Error("help should contain 'list'")
	}
	if !strings.Contains(help, "favorite") && !strings.Contains(help, "Favorite") {
		t.Error("help should mention favorites")
	}
}

func TestUrlsHelp(t *testing.T) {
	help := cmdHelp(t, urlsCmd)
	if !strings.Contains(help, "urls") {
		t.Error("help should contain 'urls'")
	}
	if !strings.Contains(help, "lookup") {
		t.Error("help should mention lookup subcommand")
	}
}

// --- Photos parent command help ---

func TestPhotosCmdHelp(t *testing.T) {
	help := cmdHelp(t, photosCmd)
	if !strings.Contains(help, "photos") {
		t.Error("help should contain 'photos'")
	}
	// Should list subcommands
	for _, sub := range []string{"list", "search", "show", "delete", "rotate", "set-meta", "set-privacy", "set-tags", "add-tags", "remove-tag", "set-location", "upload", "download"} {
		if !strings.Contains(help, sub) {
			t.Errorf("photos help should list subcommand %q", sub)
		}
	}
}

// --- Galleries parent command help ---

func TestGalleriesCmdHelp(t *testing.T) {
	help := cmdHelp(t, galleriesCmd)
	if !strings.Contains(help, "galleries") {
		t.Error("help should contain 'galleries'")
	}
	if !strings.Contains(help, "list") {
		t.Error("help should list 'list' subcommand")
	}
	if !strings.Contains(help, "photos") {
		t.Error("help should list 'photos' subcommand")
	}
}

// --- Groups parent command help ---

func TestGroupsCmdHelp(t *testing.T) {
	help := cmdHelp(t, groupsCmd)
	if !strings.Contains(help, "groups") {
		t.Error("help should contain 'groups'")
	}
	if !strings.Contains(help, "list") {
		t.Error("help should list 'list' subcommand")
	}
	if !strings.Contains(help, "search") {
		t.Error("help should list 'search' subcommand")
	}
}

// --- Comments parent command help ---

func TestCommentsCmdHelp(t *testing.T) {
	help := cmdHelp(t, commentsCmd)
	if !strings.Contains(help, "comments") {
		t.Error("help should contain 'comments'")
	}
	for _, sub := range []string{"list", "add", "delete"} {
		if !strings.Contains(help, sub) {
			t.Errorf("comments help should list subcommand %q", sub)
		}
	}
}

// --- Favorites parent command help ---

func TestFavoritesCmdHelp(t *testing.T) {
	help := cmdHelp(t, favoritesCmd)
	if !strings.Contains(help, "favorites") {
		t.Error("help should contain 'favorites'")
	}
	for _, sub := range []string{"list", "add", "remove"} {
		if !strings.Contains(help, sub) {
			t.Errorf("favorites help should list subcommand %q", sub)
		}
	}
}

// --- Photos rotate validation error ---

func TestPhotosRotateInvalidDegrees(t *testing.T) {
	_, cfg := setupFakeCLI(t)

	cmd, buf := cmdContext(t, cfg, true)
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

func TestStatsPopularHelp(t *testing.T) {
	help := cmdHelp(t, statsPopularCmd)
	if !strings.Contains(help, "popular") {
		t.Error("help should contain 'popular'")
	}
	if !strings.Contains(help, "photo") && !strings.Contains(help, "Photo") {
		t.Error("help should mention photos")
	}
}

// --- URLs lookup-user requires auth ---

func TestURLsLookupUserAuthRequired(t *testing.T) {
	fake, _ := setupFakeCLI(t)
	cfg := setupUnauthedCLI(t, fake.Server.URL)

	cmd, buf := cmdContext(t, cfg, true)
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

	cmd, buf := cmdContext(t, cfg, true)
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

	cmd, buf := cmdContext(t, cfg, true)
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

	cmd, buf := cmdContext(t, cfg, true)
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{ReadOnly: true})
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

	cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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
			cmd, buf := cmdContext(t, cfg, true, &AppContext{DryRun: true})
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
