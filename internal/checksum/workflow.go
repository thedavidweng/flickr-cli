package checksum

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
)

// Tagger adds checksum machine tags to photos that don't have them yet.
// It is a deep module: the entire list→probe→download→hash→tag workflow
// sits behind a single Add call.
type Tagger struct {
	API  flickr.FlickrAPI
	HTTP *http.Client
}

// AddOptions configures a checksum add operation.
type AddOptions struct {
	HashAlgo string
	UserID   string
	Force    bool
	TmpDir   string
	Page     int
	PerPage  int
	DryRun   bool
}

// AddDetail is the per-photo result of an add operation.
type AddDetail struct {
	PhotoID string `json:"photo_id"`
	Status  string `json:"status"` // "added", "skipped", "failed", "would_add"
	Tag     string `json:"tag,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddResult is the aggregate result of a checksum add operation.
type AddResult struct {
	Added   int         `json:"added"`
	Skipped int         `json:"skipped"`
	Failed  int         `json:"failed"`
	Total   int         `json:"total"`
	Details []AddDetail `json:"details"`
	Planned bool        `json:"planned,omitempty"`
}

// Add runs the full checksum-add workflow over the user's photos.
func (t *Tagger) Add(ctx context.Context, opts AddOptions) (*AddResult, error) {
	if err := ValidateAlgorithm(opts.HashAlgo); err != nil {
		return nil, err
	}

	tmpDir := opts.TmpDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}

	listParams := map[string]string{
		"user_id":  opts.UserID,
		"page":     fmt.Sprintf("%d", opts.Page),
		"per_page": fmt.Sprintf("%d", opts.PerPage),
	}

	var listResult struct {
		Photos struct {
			Photo []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"photo"`
			Total int `json:"total"`
		} `json:"photos"`
	}

	if err := t.API.Call(ctx, "flickr.people.getPhotos", listParams, &listResult); err != nil {
		return nil, err
	}

	result := &AddResult{
		Total:   listResult.Photos.Total,
		Details: []AddDetail{},
	}

	for _, photo := range listResult.Photos.Photo {
		if !opts.Force {
			hasTag, err := photoHasChecksum(ctx, t.API, photo.ID, opts.HashAlgo)
			if err == nil && hasTag {
				result.Skipped++
				result.Details = append(result.Details, AddDetail{
					PhotoID: photo.ID,
					Status:  "skipped",
					Reason:  "checksum tag already exists",
				})
				continue
			}
		}

		if opts.DryRun {
			result.Details = append(result.Details, AddDetail{
				PhotoID: photo.ID,
				Status:  "would_add",
			})
			continue
		}

		sourceURL, err := originalSourceURL(ctx, t.API, photo.ID)
		if err != nil {
			result.Failed++
			result.Details = append(result.Details, AddDetail{
				PhotoID: photo.ID,
				Status:  "failed",
				Error:   fmt.Sprintf("getSizes: %v", err),
			})
			continue
		}

		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("flickr-checksum-%s", photo.ID))
		hash, dlErr := downloadAndHash(t.HTTP, sourceURL, tmpFile, opts.HashAlgo)
		_ = os.Remove(tmpFile)

		if dlErr != nil {
			result.Failed++
			result.Details = append(result.Details, AddDetail{
				PhotoID: photo.ID,
				Status:  "failed",
				Error:   fmt.Sprintf("checksum: %v", dlErr),
			})
			continue
		}

		machineTag := FormatMachineTag(opts.HashAlgo, hash)
		tagParams := map[string]string{
			"photo_id": photo.ID,
			"tags":     machineTag,
		}
		if err := t.API.Call(ctx, "flickr.photos.addTags", tagParams, nil); err != nil {
			result.Failed++
			result.Details = append(result.Details, AddDetail{
				PhotoID: photo.ID,
				Status:  "failed",
				Error:   fmt.Sprintf("addTags: %v", err),
			})
			continue
		}

		result.Added++
		result.Details = append(result.Details, AddDetail{
			PhotoID: photo.ID,
			Tag:     machineTag,
			Status:  "added",
		})
	}

	if opts.DryRun {
		result.Planned = true
	}

	return result, nil
}

// Verifier downloads photos and checks their checksums against stored machine tags.
type Verifier struct {
	API  flickr.FlickrAPI
	HTTP *http.Client
}

// VerifyOptions configures a checksum verify operation.
type VerifyOptions struct {
	TmpDir  string
	Page    int
	PerPage int
}

// VerifyReport is the aggregate result of a verify operation.
type VerifyReport struct {
	Summary VerifyResult       `json:"summary"`
	Results []PhotoVerifyResult `json:"results"`
}

// Verify runs the full checksum-verify workflow over photos with checksum tags.
func (v *Verifier) Verify(ctx context.Context, opts VerifyOptions) (*VerifyReport, error) {
	tmpDir := opts.TmpDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}

	searchParams := map[string]string{
		"user_id":      "me",
		"machine_tags": "checksum:*",
		"page":         fmt.Sprintf("%d", opts.Page),
		"per_page":     fmt.Sprintf("%d", opts.PerPage),
	}

	var searchResult struct {
		Photos struct {
			Photo []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"photo"`
		} `json:"photos"`
	}

	if err := v.API.Call(ctx, "flickr.photos.search", searchParams, &searchResult); err != nil {
		return nil, err
	}

	report := &VerifyReport{
		Results: []PhotoVerifyResult{},
	}

	for _, photo := range searchResult.Photos.Photo {
		pr := verifyPhoto(ctx, v.API, v.HTTP, photo.ID, tmpDir)
		report.Results = append(report.Results, pr)
		switch pr.Status {
		case VerifyValid:
			report.Summary.Valid++
		case VerifyMissing:
			report.Summary.Missing++
		case VerifyMismatch:
			report.Summary.Mismatch++
		case VerifyFailed:
			report.Summary.Failed++
		}
	}

	return report, nil
}

// photoHasChecksum checks whether a photo already has a checksum machine tag
// for the given algorithm.
func photoHasChecksum(ctx context.Context, api flickr.FlickrAPI, photoID, algorithm string) (bool, error) {
	tags, err := getPhotoTags(ctx, api, photoID)
	if err != nil {
		return false, err
	}
	for _, tag := range tags {
		algo, _ := ParseMachineTag(tag)
		if algo == algorithm {
			return true, nil
		}
	}
	return false, nil
}

// getPhotoTags fetches the raw tag strings for a photo.
func getPhotoTags(ctx context.Context, api flickr.FlickrAPI, photoID string) ([]string, error) {
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
	if err := api.Call(ctx, "flickr.photos.getInfo", infoParams, &infoResult); err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(infoResult.Photo.Tags.Tag))
	for _, t := range infoResult.Photo.Tags.Tag {
		tags = append(tags, t.Raw)
	}
	return tags, nil
}

// originalSourceURL returns the source URL for the original size of a photo.
func originalSourceURL(ctx context.Context, api flickr.FlickrAPI, photoID string) (string, error) {
	sizes, err := api.GetSizes(ctx, photoID)
	if err != nil {
		return "", err
	}
	for _, s := range sizes {
		if s.Label == "Original" {
			return s.Source, nil
		}
	}
	if len(sizes) > 0 {
		return sizes[len(sizes)-1].Source, nil
	}
	return "", fmt.Errorf("no download URL available")
}

// downloadAndHash downloads a file from url to tmpPath and computes its hash.
func downloadAndHash(httpClient *http.Client, url, tmpPath, algorithm string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("downloading: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	_ = f.Close()

	return FileHash(tmpPath, algorithm)
}

// verifyPhoto downloads a photo and verifies its checksum against the stored tag.
func verifyPhoto(ctx context.Context, api flickr.FlickrAPI, httpClient *http.Client, photoID, tmpDir string) PhotoVerifyResult {
	tags, err := getPhotoTags(ctx, api, photoID)
	if err != nil {
		return PhotoVerifyResult{
			PhotoID: photoID,
			Status:  VerifyFailed,
			Error:   fmt.Sprintf("getInfo: %v", err),
		}
	}

	var algorithm, expectedHash string
	for _, tag := range tags {
		algo, val := ParseMachineTag(tag)
		if algo != "" {
			algorithm = algo
			expectedHash = val
			break
		}
	}

	if expectedHash == "" {
		return PhotoVerifyResult{
			PhotoID: photoID,
			Status:  VerifyMissing,
		}
	}

	sourceURL, err := originalSourceURL(ctx, api, photoID)
	if err != nil {
		return PhotoVerifyResult{
			PhotoID:  photoID,
			Status:   VerifyFailed,
			Expected: expectedHash,
			Error:    fmt.Sprintf("getSizes: %v", err),
		}
	}

	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("flickr-verify-%s", photoID))
	actualHash, dlErr := downloadAndHash(httpClient, sourceURL, tmpFile, algorithm)
	_ = os.Remove(tmpFile)

	if dlErr != nil {
		return PhotoVerifyResult{
			PhotoID:  photoID,
			Status:   VerifyFailed,
			Expected: expectedHash,
			Error:    fmt.Sprintf("checksum: %v", dlErr),
		}
	}

	if actualHash == expectedHash {
		return PhotoVerifyResult{
			PhotoID:  photoID,
			Status:   VerifyValid,
			Expected: expectedHash,
			Actual:   actualHash,
		}
	}

	return PhotoVerifyResult{
		PhotoID:  photoID,
		Status:   VerifyMismatch,
		Expected: expectedHash,
		Actual:   actualHash,
	}
}
