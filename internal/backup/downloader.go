package backup

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

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
	Media            string // "photo" or "video" (from extras)
	OriginalFormat   string // e.g. "jpg", "png" (from extras)
}

// DownloadOptions configures the downloader.
type DownloadOptions struct {
	Force    bool
	Size     string
	SizeMax  int
	Exif     bool
	Metadata string
}

// DownloadSummary is the result of a download operation.
type DownloadSummary struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// downloadResult indicates whether a file was downloaded or skipped.
type downloadResult int

const (
	downloadCompleted downloadResult = iota
	downloadSkipped
)

// Downloader handles downloading photos.
type Downloader struct {
	HTTP        *http.Client
	Client      flickr.FlickrAPI
	Concurrency int
	Events      *output.EventWriter
}

// Download downloads photos and metadata concurrently.
func (d *Downloader) Download(ctx context.Context, items []DownloadItem, opts DownloadOptions) (*DownloadSummary, error) {
	summary := &DownloadSummary{
		Total: len(items),
	}

	workers := d.Concurrency
	if workers < 1 {
		workers = 1
	}

	ch := make(chan *DownloadItem, len(items))
	for i := range items {
		ch <- &items[i]
	}
	close(ch)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				result, err := d.downloadItem(ctx, item, opts)
				if err != nil {
					mu.Lock()
					summary.Failed++
					mu.Unlock()
					d.Events.Emit(&model.Event{
						Type:    "download_failed",
						PhotoID: item.PhotoID,
						Message: err.Error(),
					})
					continue
				}
				mu.Lock()
				if result == downloadSkipped {
					summary.Skipped++
				} else {
					summary.Completed++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return summary, nil
}

func (d *Downloader) downloadItem(ctx context.Context, item *DownloadItem, opts DownloadOptions) (downloadResult, error) {
	if !opts.Force {
		if _, err := os.Stat(item.FilePath); err == nil {
			return downloadSkipped, nil
		}
	}

	// Resolve download URL: for videos use getStreamInfo, for photos use getSizes
	var downloadURL string
	var media string

	if item.Media == "video" {
		streams, err := d.Client.GetVideoStreams(ctx, item.PhotoID)
		if err == nil && len(streams) > 0 {
			best, err := flickr.SelectBestStream(streams)
			if err == nil {
				downloadURL = best.Source
				media = "video"
			}
		}
	}

	if downloadURL == "" {
		sizes, err := d.Client.GetSizes(ctx, item.PhotoID)
		if err != nil {
			return downloadCompleted, fmt.Errorf("getting sizes: %w", err)
		}

		var size flickr.Size
		if opts.SizeMax > 0 {
			size, err = flickr.SelectSizeByMaxDimension(sizes, opts.SizeMax)
		} else {
			size, err = flickr.SelectSize(sizes, opts.Size)
		}
		if err != nil {
			return downloadCompleted, fmt.Errorf("selecting size: %w", err)
		}
		downloadURL = size.Source
		media = size.Media
	}

	// Fix file extension based on actual download URL and media type
	actualExt := flickr.DeriveExtension(downloadURL, media, item.OriginalFormat)
	item.FilePath = replaceExt(item.FilePath, actualExt)
	if item.MetadataPathJSON != "" {
		item.MetadataPathJSON = item.FilePath + ".json"
	}
	if item.MetadataPathYAML != "" {
		item.MetadataPathYAML = item.FilePath + ".yaml"
	}

	dir := filepath.Dir(item.FilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return downloadCompleted, fmt.Errorf("creating dir: %w", err)
	}

	resp, err := d.HTTP.Get(downloadURL)
	if err != nil {
		return downloadCompleted, fmt.Errorf("downloading: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return downloadCompleted, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpPath := item.FilePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return downloadCompleted, fmt.Errorf("creating file: %w", err)
	}

	h := md5.New()
	writer := io.MultiWriter(f, h)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return downloadCompleted, fmt.Errorf("writing file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return downloadCompleted, err
	}

	if err := os.Rename(tmpPath, item.FilePath); err != nil {
		return downloadCompleted, err
	}

	// Write metadata sidecars if requested.
	if item.MetadataPathJSON != "" || item.MetadataPathYAML != "" {
		d.writeSidecars(ctx, item, opts.Exif)
	}

	return downloadCompleted, nil
}

// replaceExt replaces the file extension in a path.
func replaceExt(filePath, newExt string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return filePath + "." + newExt
	}
	return filePath[:len(filePath)-len(ext)] + "." + newExt
}

func (d *Downloader) writeSidecars(ctx context.Context, item *DownloadItem, includeExif bool) {
	params := map[string]string{"photo_id": item.PhotoID}
	var info map[string]any
	if err := d.Client.Call(ctx, "flickr.photos.getInfo", params, &info); err != nil {
		return
	}

	// Optionally fetch and include EXIF data
	if includeExif {
		if exifData, err := d.Client.GetExif(ctx, item.PhotoID); err == nil {
			info["exif"] = exifData
		}
	}

	// Clean up Flickr's {"_content": "value"} pattern for cleaner sidecars
	if cleaned, ok := flickr.CleanContent(info).(map[string]any); ok {
		info = cleaned
	}

	if item.MetadataPathJSON != "" {
		if data, err := json.MarshalIndent(info, "", "  "); err == nil {
			_ = os.WriteFile(item.MetadataPathJSON, data, 0o644)
		}
	}
	if item.MetadataPathYAML != "" {
		if data, err := yaml.Marshal(info); err == nil {
			_ = os.WriteFile(item.MetadataPathYAML, data, 0o644)
		}
	}
}
