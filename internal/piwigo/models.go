package piwigo

// ImportOptions configures the Piwigo import.
type ImportOptions struct {
	URL         string
	Username    string
	Password    string
	AlbumPrefix string
	ImportAlbum string
	Dedupe      string
	Hash        string
	Limit       int
	Resume      bool
}

// ImportSummary is the result of a Piwigo import.
type ImportSummary struct {
	Planned   int `json:"planned"`
	Succeeded int `json:"succeeded"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}
