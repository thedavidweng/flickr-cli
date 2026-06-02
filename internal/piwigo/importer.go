package piwigo

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
	"github.com/thedavidweng/flickr-cli/internal/safety"
)

// Importer runs the Piwigo import.
type Importer struct {
	Events    output.EventWriter
	Gate      safety.GateInput
	Profile   string
	RequestID string
	Flickr    *flickr.Client
}

// Import runs the Piwigo import.
func (i *Importer) Import(ctx context.Context, opts ImportOptions) (*ImportSummary, error) {
	gateResult := safety.Check(i.Gate, safety.Mutation{
		Command: "piwigo.import",
		Method:  "flickr.upload",
		Risk:    safety.RiskMediumWrite,
	})

	if gateResult.Error != nil {
		return nil, fmt.Errorf("%s", gateResult.Error.Message)
	}

	if gateResult.Planned {
		return &ImportSummary{Planned: 0}, nil
	}

	// Create Piwigo client
	piwigo := NewClient(opts.URL, opts.Username, opts.Password)

	// Login to Piwigo
	if err := piwigo.Login(ctx); err != nil {
		return nil, fmt.Errorf("piwigo login: %w", err)
	}

	// Get categories
	categories, err := piwigo.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting categories: %w", err)
	}

	summary := &ImportSummary{}

	// Process each category
	for _, cat := range categories {
		if cat.NbImages == 0 {
			continue
		}

		// Get images in this category
		page := 1
		perPage := 100
		for {
			images, totalPages, err := piwigo.GetCategoryImages(ctx, cat.ID, page, perPage)
			if err != nil {
				return nil, fmt.Errorf("getting images for category %s: %w", cat.ID, err)
			}

			for _, img := range images {
				summary.Planned++

				// Check limit
				if opts.Limit > 0 && summary.Succeeded >= opts.Limit {
					return summary, nil
				}

				// Check deduplication
				if opts.Dedupe == "checksum" && img.MD5Sum != "" {
					exists, err := piwigo.ImageExists(ctx, []string{img.MD5Sum})
					if err == nil && exists[img.MD5Sum] {
						summary.Skipped++
						i.Events.Emit(model.Event{
							Type:    "import_skipped",
							PhotoID: img.ID,
							Message: "already exists (checksum match)",
						})
						continue
					}
				}

				// Download image from Piwigo
				imageURL := fmt.Sprintf("%s/upload/%s", opts.URL, img.File)
				tmpFile, err := downloadToTemp(ctx, imageURL)
				if err != nil {
					summary.Failed++
					i.Events.Emit(model.Event{
						Type:    "import_failed",
						PhotoID: img.ID,
						Message: fmt.Sprintf("download failed: %v", err),
					})
					continue
				}
				defer os.Remove(tmpFile)

				// Build tags
				tags := Tags(&img)

				// Build albums
				albums := Albums(&img, categories, opts.AlbumPrefix, opts.ImportAlbum)

				// Upload to Flickr
				uploadOpts := flickr.UploadOptions{
					Title:       img.Name,
					Description: img.Comment,
					Tags:        tags,
				}

				result, err := i.Flickr.Upload(ctx, tmpFile, uploadOpts)
				if err != nil {
					summary.Failed++
					i.Events.Emit(model.Event{
						Type:    "import_failed",
						PhotoID: img.ID,
						Message: fmt.Sprintf("upload failed: %v", err),
					})
					continue
				}

				// Add to albums
				for _, albumName := range albums {
					if err := i.Flickr.AddToAlbum(ctx, albumName, result.PhotoID); err != nil {
						i.Events.Emit(model.Event{
							Type:    "import_warning",
							PhotoID: img.ID,
							Message: fmt.Sprintf("failed to add to album %s: %v", albumName, err),
						})
					}
				}

				summary.Succeeded++
				i.Events.Emit(model.Event{
					Type:    "import_success",
					PhotoID: img.ID,
					Message: fmt.Sprintf("uploaded as %s", result.PhotoID),
				})
			}

			if page >= totalPages {
				break
			}
			page++
		}
	}

	return summary, nil
}

// downloadToTemp downloads a URL to a temporary file.
func downloadToTemp(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Calculate MD5 while downloading
	h := md5.New()
	reader := io.TeeReader(resp.Body, h)

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "piwigo-*.jpg")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}
