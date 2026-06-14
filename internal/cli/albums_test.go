package cli

import (
	"bytes"
	"testing"
)

func TestAlbumsHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"albums", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestAlbumsListHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"albums", "list", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}
