package piwigo

import "testing"

func TestPrivacyFromLevel(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{0, "public"},
		{1, "friends-family"},
		{2, "friends-family"},
		{4, "friends-family"},
		{8, "private"},
		{100, "private"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := PrivacyFromLevel(tt.level)
			if got != tt.expected {
				t.Errorf("PrivacyFromLevel(%d) = %q, want %q", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLocalPath(t *testing.T) {
	tests := []struct {
		root     string
		piwigo   string
		expected string
	}{
		{"/uploads", "./upload/photos/image.jpg", "/uploads/photos/image.jpg"},
		{"/uploads", "upload/photos/image.jpg", "/uploads/photos/image.jpg"},
		{"/uploads", "/upload/photos/image.jpg", "/uploads/photos/image.jpg"},
		{"/uploads", "photos/image.jpg", "/uploads/photos/image.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.piwigo, func(t *testing.T) {
			got := LocalPath(tt.root, tt.piwigo)
			if got != tt.expected {
				t.Errorf("LocalPath(%q, %q) = %q, want %q", tt.root, tt.piwigo, got, tt.expected)
			}
		})
	}
}

func TestTags(t *testing.T) {
	record := ImageRecord{
		Tags: []string{"nature", "sunset"},
	}

	tags := Tags(record, "md5", "abc123")
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[0] != "nature" {
		t.Errorf("expected nature, got %s", tags[0])
	}
	if tags[2] != "checksum:md5=abc123" {
		t.Errorf("expected checksum tag, got %s", tags[2])
	}

	// Without hash
	tags = Tags(record, "", "")
	if len(tags) != 2 {
		t.Errorf("expected 2 tags without hash, got %d", len(tags))
	}
}

func TestAlbums(t *testing.T) {
	record := ImageRecord{
		Categories: []string{"Vacation", "2024"},
	}

	albums := Albums(record, "", "Imported")
	if len(albums) != 3 {
		t.Fatalf("expected 3 albums, got %d", len(albums))
	}
	if albums[0] != "Imported" {
		t.Errorf("expected Imported first, got %s", albums[0])
	}

	// With prefix
	albums = Albums(record, "Piwigo/", "")
	if len(albums) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(albums))
	}
	if albums[0] != "Piwigo/Vacation" {
		t.Errorf("expected Piwigo/Vacation, got %s", albums[0])
	}

	// Empty category
	record.Categories = []string{"", "Valid"}
	albums = Albums(record, "", "Import")
	if len(albums) != 2 {
		t.Errorf("expected 2 albums (skipping empty), got %d", len(albums))
	}
}
