package cli

import (
	"runtime"

	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
)

// These are set at build time via ldflags.
var (
	Version       = "dev"
	Commit        = "unknown"
	Date          = "unknown"
	SchemaVersion = "2026-06-02"
)

// VersionData is the data for the version command JSON output.
type VersionData struct {
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	Date          string `json:"date"`
	GoVersion     string `json:"go_version"`
	SchemaVersion string `json:"schema_version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := output.Renderer{
			Out:    cmd.OutOrStdout(),
			Err:    cmd.ErrOrStderr(),
			JSON:   app.JSON,
			Pretty: app.Pretty,
		}

		data := VersionData{
			Version:       Version,
			Commit:        Commit,
			Date:          Date,
			GoVersion:     runtime.Version(),
			SchemaVersion: SchemaVersion,
		}

		if app.JSON {
			return r.Success(output.RuntimeMetaInput{
				Command:   "version",
				Profile:   app.Profile,
				RequestID: app.RequestID,
				StartedAt: app.StartedAt,
			}, data, nil)
		}

		r.Human("flickr version %s (commit: %s, date: %s)\n", Version, Commit, Date)
		return nil
	},
}
