package cli

import "testing"

func TestPhotosHelp(t *testing.T) {
	help := subcommandHelp(t, rootCmd, "photos")
	if len(help) == 0 {
		t.Error("expected help output")
	}
}

func TestPhotosListHelp(t *testing.T) {
	help := subcommandHelp(t, photosCmd, "list")
	if len(help) == 0 {
		t.Error("expected help output")
	}
}

func TestPhotosSearchHelp(t *testing.T) {
	help := subcommandHelp(t, photosCmd, "search")
	if len(help) == 0 {
		t.Error("expected help output")
	}
}

func TestPhotosListHasPageFlags(t *testing.T) {
	flags := photosListCmd.Flags()

	pageFlag := flags.Lookup("page")
	if pageFlag == nil {
		t.Fatal("expected --page flag to be registered on photosListCmd")
	}
	if pageFlag.DefValue != "1" {
		t.Errorf("expected --page default=1, got %s", pageFlag.DefValue)
	}

	perPageFlag := flags.Lookup("per-page")
	if perPageFlag == nil {
		t.Fatal("expected --per-page flag to be registered on photosListCmd")
	}
	if perPageFlag.DefValue != "50" {
		t.Errorf("expected --per-page default=50, got %s", perPageFlag.DefValue)
	}
}
