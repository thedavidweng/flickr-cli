package piwigo

// PiwigoImage represents an image from the Piwigo database.
type PiwigoImage struct {
	ID        int64    `db:"id"`
	Path      string   `db:"path"`
	Name      string   `db:"name"`
	Comment   string   `db:"comment"`
	MD5Sum    string   `db:"md5sum"`
	Level     int      `db:"level"`
	Latitude  *float64 `db:"latitude"`
	Longitude *float64 `db:"longitude"`
}

// PiwigoTag represents a tag from the Piwigo database.
type PiwigoTag struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// PiwigoCategory represents a category from the Piwigo database.
type PiwigoCategory struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	UpperCat int64  `db:"uppercats"`
}

// ImageRecord is a complete image record with tags and categories.
type ImageRecord struct {
	Image      PiwigoImage
	Tags       []string
	Categories []string
}
