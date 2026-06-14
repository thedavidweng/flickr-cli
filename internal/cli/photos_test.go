package cli

import (
	"bytes"
	"testing"
)

func TestPhotosHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"photos", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestPhotosListHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"photos", "list", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestPhotosSearchHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"photos", "search", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
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
