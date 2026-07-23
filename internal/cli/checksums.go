package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/flickr-cli/internal/checksum"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

var checksumsCmd = &cobra.Command{
	Use:   "checksums",
	Short: "Manage photo checksums",
}

var checksumsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add checksum machine tags to photos",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "checksums.add",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		if app.ReadOnly {
			return r.Failure(meta, output.ErrorWithDetails(
				model.ErrReadOnlyViolation,
				"Operation blocked by --read-only flag",
				map[string]any{"command": "checksums.add", "flag": "--read-only"},
			))
		}

		hashAlgo, _ := cmd.Flags().GetString("hash")
		userID, _ := cmd.Flags().GetString("user-id")
		force, _ := cmd.Flags().GetBool("force")
		tmpDir, _ := cmd.Flags().GetString("tmp-dir")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		tagger := &checksum.Tagger{
			API:  client,
			HTTP: client.HTTP,
		}

		result, err := tagger.Add(cmd.Context(), checksum.AddOptions{
			HashAlgo: hashAlgo,
			UserID:   userID,
			Force:    force,
			TmpDir:   tmpDir,
			Page:     page,
			PerPage:  perPage,
			DryRun:   app.DryRun,
		})
		if err != nil {
			if err := checksum.ValidateAlgorithm(hashAlgo); err != nil {
				return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
			}
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var warnings []string
		if result.Skipped > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) already had checksum tags", result.Skipped))
		}
		if result.Failed > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) failed", result.Failed))
		}

		return r.Success(meta, result, warnings)
	},
}

var checksumsVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify checksums against original files",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "checksums.verify",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}
		if err := requireAuth(&r, meta, client); err != nil {
			return err
		}

		tmpDir, _ := cmd.Flags().GetString("tmp-dir")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		verifier := &checksum.Verifier{
			API:  client,
			HTTP: client.HTTP,
		}

		report, err := verifier.Verify(cmd.Context(), checksum.VerifyOptions{
			TmpDir:  tmpDir,
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var warnings []string
		if report.Summary.Mismatch > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) have mismatched checksums", report.Summary.Mismatch))
		}
		if report.Summary.Failed > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) could not be verified", report.Summary.Failed))
		}

		return r.Success(meta, report, warnings)
	},
}

var checksumsSearchCmd = &cobra.Command{
	Use:   "search [checksum]",
	Short: "Search photos by checksum",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := GetAppContext(cmd.Context())
		r := newRenderer(app, cmd)
		meta := output.RuntimeMetaInput{
			Command:   "checksums.search",
			Profile:   app.Profile,
			RequestID: app.RequestID,
			StartedAt: app.StartedAt,
		}

		client, _, err := getClient(app)
		if err != nil {
			return r.Failure(meta, output.Errorf(model.ErrConfig, "%v", err))
		}

		checksumValue := args[0]
		userID, _ := cmd.Flags().GetString("user-id")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		params := map[string]string{
			"machine_tags": fmt.Sprintf("checksum:*=%s", checksumValue),
			"page":         fmt.Sprintf("%d", page),
			"per_page":     fmt.Sprintf("%d", perPage),
		}
		if userID != "" {
			params["user_id"] = userID
		}

		var result struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Owner string `json:"owner"`
					Tags  string `json:"tags"`
				} `json:"photo"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.photos.search", params, &result); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		photos := make([]model.PhotoSummary, len(result.Photos.Photo))
		for i, p := range result.Photos.Photo {
			photos[i] = model.PhotoSummary{
				ID:    p.ID,
				Title: p.Title,
				Owner: p.Owner,
			}
		}

		if app.JSON {
			return r.Success(meta, map[string]any{
				"items": photos,
				"pagination": map[string]any{
					"page":     result.Photos.Page,
					"pages":    result.Photos.Pages,
					"per_page": result.Photos.PerPage,
					"total":    result.Photos.Total,
				},
			}, nil)
		}

		tw := output.NewTableWriter(r.Out)
		tw.Header("ID", "Title", "Owner")
		for _, p := range photos {
			tw.Row(p.ID, p.Title, p.Owner)
		}
		return tw.Flush()
	},
}

func init() {
	checksumsAddCmd.Flags().String("hash", "md5", "hash algorithm: md5|sha1")
	checksumsAddCmd.Flags().String("user-id", "me", "user ID or 'me'")
	checksumsAddCmd.Flags().Bool("force", false, "recompute even when tag exists")
	checksumsAddCmd.Flags().String("tmp-dir", "", "temporary directory for downloads")
	checksumsAddCmd.Flags().Int("page", 1, "page number")
	checksumsAddCmd.Flags().Int("per-page", 50, "items per page")

	checksumsVerifyCmd.Flags().String("tmp-dir", "", "temporary directory for downloads")
	checksumsVerifyCmd.Flags().Int("page", 1, "page number")
	checksumsVerifyCmd.Flags().Int("per-page", 50, "items per page")

	checksumsSearchCmd.Flags().String("user-id", "", "user ID or 'me'")
	checksumsSearchCmd.Flags().Int("page", 1, "page number")
	checksumsSearchCmd.Flags().Int("per-page", 50, "items per page")

	checksumsCmd.AddCommand(checksumsAddCmd)
	checksumsCmd.AddCommand(checksumsVerifyCmd)
	checksumsCmd.AddCommand(checksumsSearchCmd)
}
