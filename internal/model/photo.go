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

// Photo represents a Flickr photo with full metadata.
type Photo struct {
	ID          string     `json:"id"`
	Secret      string     `json:"secret"`
	Server      string     `json:"server"`
	Farm        int        `json:"farm"`
	Owner       User       `json:"owner"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Visibility  Visibility `json:"visibility"`
	Dates       Dates      `json:"dates"`
	Tags        []Tag      `json:"tags"`
	Location    *Location  `json:"location,omitempty"`
	URLs        PhotoURLs  `json:"urls"`
}

// Visibility represents photo visibility settings.
type Visibility struct {
	IsPublic bool `json:"is_public"`
	IsFriend bool `json:"is_friend"`
	IsFamily bool `json:"is_family"`
}

// Dates represents photo date information.
type Dates struct {
	Posted           string `json:"posted"`
	Taken            string `json:"taken"`
	TakenGranularity int    `json:"taken_granularity"`
	TakenUnknown     bool   `json:"taken_unknown"`
	LastUpdate       string `json:"last_update"`
	Uploaded         string `json:"uploaded"`
}

// Location represents geo-location data.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  int     `json:"accuracy"`
}

// PhotoURLs contains photo page URLs.
type PhotoURLs struct {
	HTML     string `json:"html"`
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
	Original string `json:"original"`
}
