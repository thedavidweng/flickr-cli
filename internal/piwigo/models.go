package piwigo

// ImportOptions configures the Piwigo import.
type ImportOptions struct {
	URL         string
	Username    string
	Password    string
	AlbumPrefix string
	ImportAlbum string
	Dedupe      string
	Limit       int
}

// ImportSummary is the result of a Piwigo import.
type ImportSummary struct {
	Planned   int `json:"planned"`
	Succeeded int `json:"succeeded"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// ImportPlan is the read-only preview produced in dry-run mode.
type ImportPlan struct {
	PlannedPhotos int `json:"planned_photos"`
	PlannedAlbums int `json:"planned_albums"`
	Skipped       int `json:"skipped"`
}
