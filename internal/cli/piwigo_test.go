package cli

import (
	"bytes"
	"testing"

	"github.com/thedavidweng/flickr-cli/internal/model"
)

func TestPiwigoHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"piwigo", "--help"})
	rootCmd.Execute()

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
