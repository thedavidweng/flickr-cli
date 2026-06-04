package model

// PhotoSummary represents a minimal photo for list views.
type PhotoSummary struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Owner       string   `json:"owner,omitempty"`
	Secret      string   `json:"secret,omitempty"`
	Server      string   `json:"server,omitempty"`
	Farm        int      `json:"farm,omitempty"`
	Media       string   `json:"media,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	OriginalURL string   `json:"original_url,omitempty"`
}
