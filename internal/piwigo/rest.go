package piwigo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a Piwigo REST API client.
type Client struct {
	BaseURL  string
	HTTP     *http.Client
	Username string
	Password string
	Token    string
}

// NewClient creates a new Piwigo REST API client.
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		HTTP:     &http.Client{Timeout: 30 * time.Second},
		Username: username,
		Password: password,
	}
}

// Login authenticates with the Piwigo instance.
func (c *Client) Login(ctx context.Context) error {
	params := map[string]string{
		"username": c.Username,
		"password": c.Password,
	}

	var result struct {
		Stat   string `json:"stat"`
		Result struct {
			Token string `json:"token"`
		} `json:"result"`
		Message string `json:"message"`
	}

	if err := c.call(ctx, "pwg.session.login", params, &result); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if result.Stat != "ok" {
		return fmt.Errorf("login failed: %s", result.Message)
	}

	c.Token = result.Result.Token
	return nil
}

// Category represents a Piwigo category/album.
type Category struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	UpperID  interface{} `json:"id_uppercat"`
	NbImages int         `json:"nb_images"`
}

// GetCategories returns all categories.
func (c *Client) GetCategories(ctx context.Context) ([]Category, error) {
	params := map[string]string{
		"recursive":   "true",
		"tree_output": "false",
		"fullname":    "true",
	}

	var result struct {
		Stat       string     `json:"stat"`
		Categories []Category `json:"result"`
	}

	if err := c.call(ctx, "pwg.categories.getList", params, &result); err != nil {
		return nil, fmt.Errorf("getting categories: %w", err)
	}

	return result.Categories, nil
}

// ImageInfo represents detailed image information from Piwigo.
type ImageInfo struct {
	ID            string  `json:"id"`
	File          string  `json:"file"`
	Name          string  `json:"name"`
	Comment       string  `json:"comment"`
	Author        string  `json:"author"`
	DateCreation  string  `json:"date_creation"`
	DateAvailable string  `json:"date_available"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	FileSize      int64   `json:"filesize"`
	Views         int     `json:"hit"`
	Rating        float64 `json:"rating_score"`
	MD5Sum        string  `json:"md5sum"`
	Level         int     `json:"level"`
	Categories    []struct {
		ID string `json:"id"`
	} `json:"categories"`
	Tags []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// GetCategoryImages returns images in a category.
func (c *Client) GetCategoryImages(ctx context.Context, categoryID string, page, perPage int) ([]ImageInfo, int, error) {
	params := map[string]string{
		"category_id": categoryID,
		"page":        fmt.Sprintf("%d", page),
		"per_page":    fmt.Sprintf("%d", perPage),
	}

	var result struct {
		Stat   string      `json:"stat"`
		Images []ImageInfo `json:"result"`
		Paging struct {
			TotalPages int `json:"total_pages"`
		} `json:"paging"`
	}

	if err := c.call(ctx, "pwg.categories.getImages", params, &result); err != nil {
		return nil, 0, fmt.Errorf("getting category images: %w", err)
	}

	return result.Images, result.Paging.TotalPages, nil
}

// GetImageInfo returns detailed information about an image.
func (c *Client) GetImageInfo(ctx context.Context, imageID string) (*ImageInfo, error) {
	params := map[string]string{
		"image_id": imageID,
	}

	var result struct {
		Stat  string    `json:"stat"`
		Image ImageInfo `json:"result"`
	}

	if err := c.call(ctx, "pwg.images.getInfo", params, &result); err != nil {
		return nil, fmt.Errorf("getting image info: %w", err)
	}

	return &result.Image, nil
}

// GetTags returns all tags.
func (c *Client) GetTags(ctx context.Context) ([]Tag, error) {
	params := map[string]string{
		"sort_by": "counter",
	}

	var result struct {
		Stat string `json:"stat"`
		Tags []Tag  `json:"result"`
	}

	if err := c.call(ctx, "pwg.tags.getList", params, &result); err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	return result.Tags, nil
}

// Tag represents a Piwigo tag.
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"counter"`
}

// ImageExists checks if images exist by their MD5 checksums.
func (c *Client) ImageExists(ctx context.Context, md5sums []string) (map[string]bool, error) {
	params := map[string]string{
		"md5sum_list": strings.Join(md5sums, ","),
	}

	var result struct {
		Stat    string          `json:"stat"`
		Results map[string]bool `json:"result"`
	}

	if err := c.call(ctx, "pwg.images.exist", params, &result); err != nil {
		return nil, fmt.Errorf("checking image existence: %w", err)
	}

	return result.Results, nil
}

// call makes a REST API call to Piwigo.
func (c *Client) call(ctx context.Context, method string, params map[string]string, result interface{}) error {
	endpoint := fmt.Sprintf("%s/ws.php?format=json", c.BaseURL)

	form := url.Values{}
	form.Set("method", method)
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.Token != "" {
		req.Header.Set("Cookie", "pwg_id="+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return nil
}
