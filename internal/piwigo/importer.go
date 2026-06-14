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
)

// Importer runs the Piwigo import.
type Importer struct {
	Events    *output.EventWriter
	Profile   string
	RequestID string
	Flickr    flickr.FlickrAPI
}

// Import runs the Piwigo import.
func (i *Importer) Import(ctx context.Context, opts *ImportOptions) (*ImportSummary, error) {
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

			for idx := range images {
				img := &images[idx]
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
						i.Events.Emit(&model.Event{
							Type:    string(StateSkippedExist),
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
					i.Events.Emit(&model.Event{
						Type:    string(StateFailed),
						PhotoID: img.ID,
						Message: fmt.Sprintf("download failed: %v", err),
					})
					continue
				}

				// Build tags
				tags := Tags(img)

				// Build albums
				albums := Albums(img, categories, opts.AlbumPrefix, opts.ImportAlbum)

				// Upload to Flickr
				uploadOpts := flickr.UploadOptions{
					Title:       img.Name,
					Description: img.Comment,
					Tags:        tags,
				}

				result, err := i.Flickr.Upload(ctx, tmpFile, &uploadOpts)
				if err != nil {
					_ = os.Remove(tmpFile)
					summary.Failed++
					i.Events.Emit(&model.Event{
						Type:    string(StateFailed),
						PhotoID: img.ID,
						Message: fmt.Sprintf("upload failed: %v", err),
					})
					continue
				}

				// Add to albums
				for _, albumName := range albums {
					if err := i.Flickr.AddToAlbum(ctx, albumName, result.PhotoID); err != nil {
						i.Events.Emit(&model.Event{
							Type:    "import_warning",
							PhotoID: img.ID,
							Message: fmt.Sprintf("failed to add to album %s: %v", albumName, err),
						})
					}
				}

				// Transfer geo-location if available
				if img.Latitude != 0 || img.Longitude != 0 {
					geoParams := map[string]string{
						"photo_id": result.PhotoID,
						"lat":      fmt.Sprintf("%f", img.Latitude),
						"lon":      fmt.Sprintf("%f", img.Longitude),
					}
					if err := i.Flickr.Call(ctx, "flickr.photos.geo.setLocation", geoParams, nil); err != nil {
						i.Events.Emit(&model.Event{
							Type:    string(StateGeoDone),
							PhotoID: img.ID,
							Message: fmt.Sprintf("geo-location transfer failed: %v", err),
						})
					}
				}

				summary.Succeeded++
				i.Events.Emit(&model.Event{
					Type:    string(StateDone),
					PhotoID: img.ID,
					Message: fmt.Sprintf("uploaded as %s", result.PhotoID),
				})
				_ = os.Remove(tmpFile)
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
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = tmpFile.Close() }()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}
