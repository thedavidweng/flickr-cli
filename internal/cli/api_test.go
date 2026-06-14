package cli

import (
	"bytes"
	"testing"
)

func TestAPIHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"api", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestAPICallHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"api", "call", "--help"})
	_ = rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}
