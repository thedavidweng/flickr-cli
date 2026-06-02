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
	cmd.Flags().String("uploads", "", "")
	cmd.Flags().String("mysql-db", "", "")
	cmd.Flags().String("mysql-host", "localhost", "")
	cmd.Flags().Int("mysql-port", 3306, "")
	cmd.Flags().String("mysql-user", "", "")
	cmd.Flags().String("mysql-password", "", "")
	cmd.Flags().String("mysql-password-env", "", "")
	cmd.Flags().String("table-prefix", "", "")
	cmd.Flags().String("album-prefix", "", "")
	cmd.Flags().String("import-album", "", "")
	cmd.Flags().String("dedupe", "", "")
	cmd.Flags().String("hash", "", "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().Bool("resume", false, "")

	// Call without setting --uploads or --mysql-db
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
