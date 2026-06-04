package flickr

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// UploadOptions contains options for uploading a photo.
type UploadOptions struct {
	Title       string
	Description string
	Tags        []string
	IsPublic    bool
	IsFriend    bool
	IsFamily    bool
	SafetyLevel int
	ContentType int
	Hidden      int
}

// UploadResult is the result of a successful upload.
type UploadResult struct {
	PhotoID string
}

// flickrUploadResponse is the XML response from the upload endpoint.
type flickrUploadResponse struct {
	XMLName xml.Name `xml:"rsp"`
	Stat    string   `xml:"stat,attr"`
	Err     struct {
		Code    int    `xml:"code,attr"`
		Message string `xml:"msg,attr"`
	} `xml:"err"`
	Photoid struct {
		ID string `xml:",chardata"`
	} `xml:"photoid"`
}

// Upload uploads a photo to Flickr.
func (c *Client) Upload(ctx context.Context, filePath string, opts UploadOptions) (*UploadResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	// Build multipart body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add the photo file
	part, err := writer.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copying file: %w", err)
	}

	// Add text fields
	title := opts.Title
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}
	writer.WriteField("title", title)

	if opts.Description != "" {
		writer.WriteField("description", opts.Description)
	}

	if len(opts.Tags) > 0 {
		writer.WriteField("tags", strings.Join(opts.Tags, " "))
	}

	writer.WriteField("is_public", boolToNum(opts.IsPublic))
	writer.WriteField("is_friend", boolToNum(opts.IsFriend))
	writer.WriteField("is_family", boolToNum(opts.IsFamily))

	if opts.SafetyLevel > 0 {
		writer.WriteField("safety_level", fmt.Sprintf("%d", opts.SafetyLevel))
	}
	if opts.ContentType > 0 {
		writer.WriteField("content_type", fmt.Sprintf("%d", opts.ContentType))
	}
	if opts.Hidden > 0 {
		writer.WriteField("hidden", fmt.Sprintf("%d", opts.Hidden))
	}

	writer.Close()

	// Build OAuth signature params (exclude the photo field)
	sigParams := make(map[string][]string)
	sigParams["title"] = []string{title}
	if opts.Description != "" {
		sigParams["description"] = []string{opts.Description}
	}
	if len(opts.Tags) > 0 {
		sigParams["tags"] = []string{strings.Join(opts.Tags, " ")}
	}
	sigParams["is_public"] = []string{boolToNum(opts.IsPublic)}
	sigParams["is_friend"] = []string{boolToNum(opts.IsFriend)}
	sigParams["is_family"] = []string{boolToNum(opts.IsFamily)}

	signer := c.Signer()
	oauthParams, err := signer.Sign("POST", c.Endpoints.Upload, sigParams)
	if err != nil {
		return nil, fmt.Errorf("signing upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoints.Upload, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", AuthorizationHeader(oauthParams))
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uploading: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var uploadResp flickrUploadResponse
	if err := xml.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if uploadResp.Stat != "ok" {
		return nil, &FlickrError{
			Method: "upload",
			Code:   uploadResp.Err.Code,
			Msg:    uploadResp.Err.Message,
			Stat:   uploadResp.Stat,
		}
	}

	if uploadResp.Photoid.ID == "" {
		return nil, fmt.Errorf("upload succeeded but no photo ID returned")
	}

	return &UploadResult{
		PhotoID: uploadResp.Photoid.ID,
	}, nil
}

// AddToAlbum adds a photo to an album/photoset.
func (c *Client) AddToAlbum(ctx context.Context, albumID, photoID string) error {
	params := map[string]string{
		"photoset_id": albumID,
		"photo_id":    photoID,
	}
	return c.Call(ctx, "flickr.photosets.addPhoto", params, nil)
}

func boolToNum(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
