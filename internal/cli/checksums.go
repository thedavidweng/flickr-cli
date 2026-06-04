package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/thedavidweng/flickr-cli/internal/checksum"
	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/spf13/cobra"
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

		if err := checksum.ValidateAlgorithm(hashAlgo); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrValidationFailed, "%v", err))
		}

		if tmpDir == "" {
			tmpDir = os.TempDir()
		}

		// List photos
		listParams := map[string]string{
			"user_id":  userID,
			"page":     fmt.Sprintf("%d", page),
			"per_page": fmt.Sprintf("%d", perPage),
		}

		var listResult struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"photo"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.people.getPhotos", listParams, &listResult); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var added, skipped, failed int
		var details []map[string]any

		for _, photo := range listResult.Photos.Photo {
			// Check if already has checksum tag (unless force)
			if !force {
				hasTag, err := photoHasChecksum(client, cmd, photo.ID, hashAlgo)
				if err == nil && hasTag {
					skipped++
					details = append(details, map[string]any{
						"photo_id": photo.ID,
						"status":   "skipped",
						"reason":   "checksum tag already exists",
					})
					continue
				}
			}

			if app.DryRun {
				details = append(details, map[string]any{
					"photo_id": photo.ID,
					"status":   "would_add",
				})
				continue
			}

			// Get sizes to find download URL
			sourceURL, err := getOriginalSourceURL(client, cmd, photo.ID)
			if err != nil {
				failed++
				details = append(details, map[string]any{
					"photo_id": photo.ID,
					"status":   "failed",
					"error":    fmt.Sprintf("getSizes: %v", err),
				})
				continue
			}

			// Download and compute checksum
			tmpFile := filepath.Join(tmpDir, fmt.Sprintf("flickr-checksum-%s", photo.ID))
			hash, dlErr := downloadAndHash(client.HTTP, sourceURL, tmpFile, hashAlgo)
			os.Remove(tmpFile)

			if dlErr != nil {
				failed++
				details = append(details, map[string]any{
					"photo_id": photo.ID,
					"status":   "failed",
					"error":    fmt.Sprintf("checksum: %v", dlErr),
				})
				continue
			}

			// Add machine tag
			machineTag := checksum.FormatMachineTag(hashAlgo, hash)
			tagParams := map[string]string{
				"photo_id": photo.ID,
				"tags":     machineTag,
			}
			if err := client.Call(cmd.Context(), "flickr.photos.addTags", tagParams, nil); err != nil {
				failed++
				details = append(details, map[string]any{
					"photo_id": photo.ID,
					"status":   "failed",
					"error":    fmt.Sprintf("addTags: %v", err),
				})
				continue
			}

			added++
			details = append(details, map[string]any{
				"photo_id": photo.ID,
				"tag":      machineTag,
				"status":   "added",
			})
		}

		data := map[string]any{
			"added":   added,
			"skipped": skipped,
			"failed":  failed,
			"total":   listResult.Photos.Total,
			"details": details,
		}
		if app.DryRun {
			data["planned"] = true
		}

		var warnings []string
		if skipped > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) already had checksum tags", skipped))
		}
		if failed > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) failed", failed))
		}

		return r.Success(meta, data, warnings)
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

		if tmpDir == "" {
			tmpDir = os.TempDir()
		}

		// Search for photos with checksum machine tags
		searchParams := map[string]string{
			"user_id":      "me",
			"machine_tags": "checksum:*",
			"page":         fmt.Sprintf("%d", page),
			"per_page":     fmt.Sprintf("%d", perPage),
		}

		var searchResult struct {
			Photos struct {
				Photo []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"photo"`
				Page    int `json:"page"`
				Pages   int `json:"pages"`
				PerPage int `json:"perpage"`
				Total   int `json:"total"`
			} `json:"photos"`
		}

		if err := client.Call(cmd.Context(), "flickr.photos.search", searchParams, &searchResult); err != nil {
			return r.Failure(meta, output.Errorf(model.ErrFlickrAPI, "%v", err))
		}

		var results []checksum.PhotoVerifyResult
		var summary checksum.VerifyResult

		for _, photo := range searchResult.Photos.Photo {
			pr := verifyPhoto(client, cmd, photo.ID, tmpDir)
			results = append(results, pr)
			switch pr.Status {
			case checksum.VerifyValid:
				summary.Valid++
			case checksum.VerifyMissing:
				summary.Missing++
			case checksum.VerifyMismatch:
				summary.Mismatch++
			case checksum.VerifyFailed:
				summary.Failed++
			}
		}

		var warnings []string
		if summary.Mismatch > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) have mismatched checksums", summary.Mismatch))
		}
		if summary.Failed > 0 {
			warnings = append(warnings, fmt.Sprintf("%d photo(s) could not be verified", summary.Failed))
		}

		return r.Success(meta, map[string]any{
			"summary": summary,
			"results": results,
		}, warnings)
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

// downloadAndHash downloads a file from url to tmpPath and computes its hash.
func downloadAndHash(httpClient *http.Client, url, tmpPath, algorithm string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	// Close before hashing so all data is flushed to disk.
	f.Close()

	return checksum.FileHash(tmpPath, algorithm)
}

// photoHasChecksum checks whether a photo already has a checksum machine tag
// for the given algorithm.
func photoHasChecksum(client *flickr.Client, cmd *cobra.Command, photoID, algorithm string) (bool, error) {
	infoParams := map[string]string{"photo_id": photoID}
	var infoResult struct {
		Photo struct {
			Tags struct {
				Tag []struct {
					Raw     string `json:"raw"`
					Machine int    `json:"machine"`
				} `json:"tag"`
			} `json:"tags"`
		} `json:"photo"`
	}
	if err := client.Call(cmd.Context(), "flickr.photos.getInfo", infoParams, &infoResult); err != nil {
		return false, err
	}
	for _, tag := range infoResult.Photo.Tags.Tag {
		algo, _ := checksum.ParseMachineTag(tag.Raw)
		if algo == algorithm {
			return true, nil
		}
	}
	return false, nil
}

// getOriginalSourceURL returns the source URL for the original size of a photo.
func getOriginalSourceURL(client *flickr.Client, cmd *cobra.Command, photoID string) (string, error) {
	sizeParams := map[string]string{"photo_id": photoID}
	var sizeResult struct {
		Sizes struct {
			Size []struct {
				Label  string `json:"label"`
				Source string `json:"source"`
			} `json:"size"`
		} `json:"sizes"`
	}
	if err := client.Call(cmd.Context(), "flickr.photos.getSizes", sizeParams, &sizeResult); err != nil {
		return "", err
	}

	for _, s := range sizeResult.Sizes.Size {
		if s.Label == "Original" {
			return s.Source, nil
		}
	}
	if len(sizeResult.Sizes.Size) > 0 {
		return sizeResult.Sizes.Size[len(sizeResult.Sizes.Size)-1].Source, nil
	}
	return "", fmt.Errorf("no download URL available")
}

// verifyPhoto downloads a photo and verifies its checksum against the stored tag.
func verifyPhoto(client *flickr.Client, cmd *cobra.Command, photoID, tmpDir string) checksum.PhotoVerifyResult {
	// Get photo info to find checksum tag
	infoParams := map[string]string{"photo_id": photoID}
	var infoResult struct {
		Photo struct {
			Tags struct {
				Tag []struct {
					Raw     string `json:"raw"`
					Machine int    `json:"machine"`
				} `json:"tag"`
			} `json:"tags"`
		} `json:"photo"`
	}
	if err := client.Call(cmd.Context(), "flickr.photos.getInfo", infoParams, &infoResult); err != nil {
		return checksum.PhotoVerifyResult{
			PhotoID: photoID,
			Status:  checksum.VerifyFailed,
			Error:   fmt.Sprintf("getInfo: %v", err),
		}
	}

	// Find checksum tag
	var algorithm, expectedHash string
	for _, tag := range infoResult.Photo.Tags.Tag {
		algo, val := checksum.ParseMachineTag(tag.Raw)
		if algo != "" {
			algorithm = algo
			expectedHash = val
			break
		}
	}

	if expectedHash == "" {
		return checksum.PhotoVerifyResult{
			PhotoID: photoID,
			Status:  checksum.VerifyMissing,
		}
	}

	// Get download URL
	sourceURL, err := getOriginalSourceURL(client, cmd, photoID)
	if err != nil {
		return checksum.PhotoVerifyResult{
			PhotoID:  photoID,
			Status:   checksum.VerifyFailed,
			Expected: expectedHash,
			Error:    fmt.Sprintf("getSizes: %v", err),
		}
	}

	// Download and compute checksum
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("flickr-verify-%s", photoID))
	actualHash, dlErr := downloadAndHash(client.HTTP, sourceURL, tmpFile, algorithm)
	os.Remove(tmpFile)

	if dlErr != nil {
		return checksum.PhotoVerifyResult{
			PhotoID:  photoID,
			Status:   checksum.VerifyFailed,
			Expected: expectedHash,
			Error:    fmt.Sprintf("checksum: %v", dlErr),
		}
	}

	if actualHash == expectedHash {
		return checksum.PhotoVerifyResult{
			PhotoID:  photoID,
			Status:   checksum.VerifyValid,
			Expected: expectedHash,
			Actual:   actualHash,
		}
	}

	return checksum.PhotoVerifyResult{
		PhotoID:  photoID,
		Status:   checksum.VerifyMismatch,
		Expected: expectedHash,
		Actual:   actualHash,
	}
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
