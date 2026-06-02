package model

// Tag represents a photo tag.
type Tag struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Machine bool   `json:"machine"`
}
