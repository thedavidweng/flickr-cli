package backup

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/thedavidweng/flickr-cli/internal/flickr"
	"github.com/thedavidweng/flickr-cli/internal/model"
	"github.com/thedavidweng/flickr-cli/internal/output"
)

// DownloadItem represents a photo to download.
type DownloadItem struct {
	PhotoID          string
	FilePath         string
	MetadataPathJSON string
	MetadataPathYAML string
	SizeLabel        string
}

// DownloadOptions configures the downloader.
type DownloadOptions struct {
	Force    bool
	Size     string
	Metadata string
}

// DownloadSummary is the result of a download operation.
type DownloadSummary struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// Downloader handles downloading photos.
type Downloader struct {
	HTTP        *http.Client
	Client      *flickr.Client
	Concurrency int
	Events      output.EventWriter
}

// Download downloads photos and metadata.
func (d *Downloader) Download(ctx context.Context, items []DownloadItem, opts DownloadOptions) (*DownloadSummary, error) {
	summary := &DownloadSummary{
		Total: len(items),
	}

	for _, item := range items {
		if err := d.downloadItem(ctx, item, opts); err != nil {
			summary.Failed++
			d.Events.Emit(model.Event{
				Type:    "download_failed",
				PhotoID: item.PhotoID,
				Message: err.Error(),
			})
			continue
		}
		summary.Completed++
	}

	return summary, nil
}

func (d *Downloader) downloadItem(ctx context.Context, item DownloadItem, opts DownloadOptions) error {
	if !opts.Force {
		if _, err := os.Stat(item.FilePath); err == nil {
			return nil
		}
	}

	sizes, err := d.Client.GetSizes(ctx, item.PhotoID)
	if err != nil {
		return fmt.Errorf("getting sizes: %w", err)
	}

	size, err := flickr.SelectSize(sizes, opts.Size)
	if err != nil {
		return fmt.Errorf("selecting size: %w", err)
	}

	dir := filepath.Dir(item.FilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating dir: %w", err)
	}

	resp, err := d.HTTP.Get(size.Source)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpPath := item.FilePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	h := md5.New()
	writer := io.MultiWriter(f, h)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, item.FilePath)
}
