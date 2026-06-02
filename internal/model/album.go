package model

// Album represents a Flickr photoset/album.
type Album struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	PhotoCount     int    `json:"photo_count"`
	PrimaryPhotoID string `json:"primary_photo_id"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
