package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

func TestVersionHuman(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if out == "" {
		t.Error("expected non-empty output")
	}
}

func TestVersionJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--json", "version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env model.Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !env.OK {
		t.Error("expected ok=true")
	}
	if env.Meta.Command != "version" {
		t.Errorf("expected command=version, got %s", env.Meta.Command)
	}
}
