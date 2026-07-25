package piwigo

import "testing"

func TestTags(t *testing.T) {
	image := &ImageInfo{
		Tags: []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			{ID: "1", Name: "nature"},
			{ID: "2", Name: "sunset"},
		},
	}

	tags := Tags(image)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "nature" {
		t.Errorf("expected nature, got %s", tags[0])
	}
	if tags[1] != "sunset" {
		t.Errorf("expected sunset, got %s", tags[1])
	}
}

func TestAlbums(t *testing.T) {
	image := &ImageInfo{
		Categories: []struct {
			ID string `json:"id"`
		}{
			{ID: "1"},
			{ID: "2"},
		},
	}

	categories := []Category{
		{ID: "1", Name: "Vacation"},
		{ID: "2", Name: "2024"},
	}

	albums := Albums(image, categories, "", "Imported")
	if len(albums) != 3 {
		t.Fatalf("expected 3 albums, got %d", len(albums))
	}
	if albums[0] != "Imported" {
		t.Errorf("expected Imported first, got %s", albums[0])
	}

	// With prefix
	albums = Albums(image, categories, "Piwigo/", "")
	if len(albums) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(albums))
	}
	if albums[0] != "Piwigo/Vacation" {
		t.Errorf("expected Piwigo/Vacation, got %s", albums[0])
	}
}
