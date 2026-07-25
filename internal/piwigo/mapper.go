package piwigo

// Tags builds the tag list for an image.
func Tags(image *ImageInfo) []string {
	tags := make([]string, 0, len(image.Tags))
	for _, tag := range image.Tags {
		tags = append(tags, tag.Name)
	}
	return tags
}

// Albums builds the album list for an image.
func Albums(image *ImageInfo, categories []Category, prefix, importAlbum string) []string {
	var albums []string

	if importAlbum != "" {
		albums = append(albums, importAlbum)
	}

	// Build category ID to name map
	catMap := make(map[string]string)
	for _, cat := range categories {
		catMap[cat.ID] = cat.Name
	}

	for _, cat := range image.Categories {
		name := catMap[cat.ID]
		if name == "" {
			continue
		}
		if prefix != "" {
			albums = append(albums, prefix+name)
		} else {
			albums = append(albums, name)
		}
	}

	return albums
}
