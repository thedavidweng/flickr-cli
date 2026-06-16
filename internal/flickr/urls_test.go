package flickr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSelectSize(t *testing.T) {
	sizes := []Size{
		{Label: "Thumbnail", Width: 100, Height: 75, Source: "thumb.jpg"},
		{Label: "Small", Width: 240, Height: 180, Source: "small.jpg"},
		{Label: "Medium", Width: 500, Height: 375, Source: "medium.jpg"},
		{Label: "Large", Width: 1024, Height: 768, Source: "large.jpg"},
		{Label: "Original", Width: 4000, Height: 3000, Source: "original.jpg"},
	}

	tests := []struct {
		wanted   string
		expected string
	}{
		{"original", "original.jpg"},
		{"large", "large.jpg"},
		{"medium", "medium.jpg"},
		{"small", "small.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.wanted, func(t *testing.T) {
			size, err := SelectSize(sizes, tt.wanted)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if size.Source != tt.expected {
				t.Errorf("SelectSize(%s) source = %s, want %s", tt.wanted, size.Source, tt.expected)
			}
		})
	}
}

func TestSelectSizeEmpty(t *testing.T) {
	_, err := SelectSize([]Size{}, "original")
	if err == nil {
		t.Error("expected error for empty sizes")
	}
}

func TestSelectSizeOriginalFallback(t *testing.T) {
	sizes := []Size{
		{Label: "Thumbnail", Width: 100, Height: 75, Source: "thumb.jpg"},
		{Label: "Large", Width: 1024, Height: 768, Source: "large.jpg"},
	}

	size, err := SelectSize(sizes, "original")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Source != "large.jpg" {
		t.Errorf("expected large.jpg fallback, got %s", size.Source)
	}
}

func TestGetSizes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"ok","sizes":{"size":[{"label":"Original","width":4000,"height":3000,"source":"original.jpg","url":"http://example.com","media":"photo"}]}}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:    "test-key",
		HTTP:      server.Client(),
		Endpoints: Endpoints{REST: server.URL + "/"},
	}

	sizes, err := client.GetSizes(context.Background(), "photo-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sizes) != 1 {
		t.Errorf("expected 1 size, got %d", len(sizes))
	}
}

func BenchmarkSelectSize(b *testing.B) {
	sizes := []Size{
		{Label: "Thumbnail", Width: 100, Height: 75, Source: "thumb.jpg"},
		{Label: "Small 320", Width: 320, Height: 240, Source: "small320.jpg"},
		{Label: "Small", Width: 240, Height: 180, Source: "small.jpg"},
		{Label: "Medium 640", Width: 640, Height: 480, Source: "medium640.jpg"},
		{Label: "Medium 800", Width: 800, Height: 600, Source: "medium800.jpg"},
		{Label: "Medium", Width: 500, Height: 375, Source: "medium.jpg"},
		{Label: "Large 1600", Width: 1600, Height: 1200, Source: "large1600.jpg"},
		{Label: "Large", Width: 1024, Height: 768, Source: "large.jpg"},
		{Label: "Original", Width: 4000, Height: 3000, Source: "original.jpg"},
	}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = SelectSize(sizes, "large")
	}
}

func BenchmarkDecodeShortURL(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = DecodeShortURL("https://flic.kr/p/2oFnQhB")
	}
}

func BenchmarkDeriveExtension(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = DeriveExtension("https://live.staticflickr.com/65535/54321098765_abcdef1234_b.jpg", "photo", "")
	}
}

// ---------------------------------------------------------------------------
// SelectBestStream
// ---------------------------------------------------------------------------

func TestSelectBestStream(t *testing.T) {
	t.Run("empty slice returns error", func(t *testing.T) {
		_, err := SelectBestStream(nil)
		if err == nil {
			t.Fatal("expected error for nil streams")
		}
		_, err = SelectBestStream([]VideoStream{})
		if err == nil {
			t.Fatal("expected error for empty streams")
		}
	})

	t.Run("single stream returns it", func(t *testing.T) {
		streams := []VideoStream{{Type: "720p", Width: 1280, Height: 720, Source: "720.mp4"}}
		got, err := SelectBestStream(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Source != "720.mp4" {
			t.Errorf("got source %q, want %q", got.Source, "720.mp4")
		}
	})

	t.Run("prefers orig over 1080p over 720p", func(t *testing.T) {
		streams := []VideoStream{
			{Type: "720p", Source: "720.mp4"},
			{Type: "orig", Source: "orig.mp4"},
			{Type: "1080p", Source: "1080.mp4"},
		}
		got, err := SelectBestStream(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Type != "orig" {
			t.Errorf("got type %q, want %q", got.Type, "orig")
		}
		if got.Source != "orig.mp4" {
			t.Errorf("got source %q, want %q", got.Source, "orig.mp4")
		}
	})

	t.Run("unknown types sorted after known", func(t *testing.T) {
		streams := []VideoStream{
			{Type: "4k", Source: "4k.mp4"},
			{Type: "1080p", Source: "1080.mp4"},
			{Type: "720p", Source: "720.mp4"},
		}
		got, err := SelectBestStream(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Type != "1080p" {
			t.Errorf("got type %q, want %q (known beats unknown)", got.Type, "1080p")
		}
	})

	t.Run("mix of known and unknown picks highest known", func(t *testing.T) {
		streams := []VideoStream{
			{Type: "webm", Source: "webm.mp4"},
			{Type: "720p", Source: "720.mp4"},
			{Type: "orig", Source: "orig.mp4"},
			{Type: "hls", Source: "hls.m3u8"},
		}
		got, err := SelectBestStream(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Type != "orig" {
			t.Errorf("got type %q, want %q", got.Type, "orig")
		}
	})

	t.Run("all unknown picks first alphabetically-stable", func(t *testing.T) {
		streams := []VideoStream{
			{Type: "webm", Source: "webm.mp4"},
			{Type: "hls", Source: "hls.m3u8"},
		}
		got, err := SelectBestStream(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Both have priority 99; sort is stable-ish, either is fine.
		if got.Type != "webm" && got.Type != "hls" {
			t.Errorf("got unexpected type %q", got.Type)
		}
	})

	t.Run("prefers 360p over 288p", func(t *testing.T) {
		streams := []VideoStream{
			{Type: "288p", Source: "288.mp4"},
			{Type: "360p", Source: "360.mp4"},
		}
		got, err := SelectBestStream(streams)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Type != "360p" {
			t.Errorf("got type %q, want %q", got.Type, "360p")
		}
	})
}

// ---------------------------------------------------------------------------
// selectByCode (tested through SelectSize with size codes)
// ---------------------------------------------------------------------------

func TestSelectByCode(t *testing.T) {
	sizes := []Size{
		{Label: "Square 75", Width: 75, Height: 75, Source: "sq75.jpg"},
		{Label: "Large Square", Width: 150, Height: 150, Source: "sq150.jpg"},
		{Label: "Thumbnail", Width: 100, Height: 75, Source: "thumb.jpg"},
		{Label: "Small 240", Width: 240, Height: 180, Source: "s240.jpg"},
		{Label: "Small 320", Width: 320, Height: 240, Source: "s320.jpg"},
		{Label: "Medium 500", Width: 500, Height: 375, Source: "m500.jpg"},
		{Label: "Medium 640", Width: 640, Height: 480, Source: "m640.jpg"},
		{Label: "Medium 800", Width: 800, Height: 600, Source: "m800.jpg"},
		{Label: "Large 1024", Width: 1024, Height: 768, Source: "l1024.jpg"},
		{Label: "Large 1600", Width: 1600, Height: 1200, Source: "l1600.jpg"},
		{Label: "Large 2048", Width: 2048, Height: 1536, Source: "l2048.jpg"},
		{Label: "Original", Width: 4000, Height: 3000, Source: "orig.jpg"},
	}

	tests := []struct {
		name     string
		code     string
		wantSrc  string
		wantDesc string
	}{
		{"original by code", "o", "orig.jpg", "Original label match"},
		{"2048 by label", "k", "l2048.jpg", "label contains 2048"},
		{"1600 by label", "h", "l1600.jpg", "label contains 1600"},
		{"Large by label", "l", "sq150.jpg", "label contains Large (Large Square matches first)"},
		{"640 by label", "z", "m640.jpg", "label contains 640"},
		{"Medium by label", "m", "m500.jpg", "label contains Medium (first match)"},
		{"Square by label", "s", "sq75.jpg", "label contains Square (first match)"},
		{"Large Square by label", "q", "sq150.jpg", "label contains Large Square"},
		{"Thumbnail by label", "t", "thumb.jpg", "label contains Thumbnail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectSize(sizes, tt.code)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Source != tt.wantSrc {
				t.Errorf("SelectSize(%q) source = %q, want %q (%s)", tt.code, got.Source, tt.wantSrc, tt.wantDesc)
			}
		})
	}
}

func TestSelectByCodeDimensionMatch(t *testing.T) {
	// Sizes without any matching label text for "800" code
	sizes := []Size{
		{Label: "Custom A", Width: 600, Height: 400, Source: "a.jpg"},
		{Label: "Custom B", Width: 799, Height: 500, Source: "b.jpg"},
		{Label: "Custom C", Width: 800, Height: 600, Source: "c.jpg"},
		{Label: "Custom D", Width: 1024, Height: 768, Source: "d.jpg"},
	}

	// Code "c" has labelContains "800" but none of these labels contain "800",
	// so it falls back to dimension match: largest size with width <= 800.
	got, err := SelectSize(sizes, "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Source != "c.jpg" {
		t.Errorf("dimension match: got %q, want %q", got.Source, "c.jpg")
	}
}

func TestSelectByCodeOriginalFallback(t *testing.T) {
	// No "Original" label; code "o" should fall back to the last (largest) size.
	sizes := []Size{
		{Label: "Small", Width: 240, Height: 180, Source: "small.jpg"},
		{Label: "Large", Width: 1024, Height: 768, Source: "large.jpg"},
	}

	got, err := SelectSize(sizes, "o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Source != "large.jpg" {
		t.Errorf("fallback: got %q, want %q", got.Source, "large.jpg")
	}
}

func TestSelectByCodeClosestWidthFallback(t *testing.T) {
	// Code "k" (maxDim 2048), no label contains "2048",
	// and nothing fits in 2048 (all are larger). Should pick closest by width.
	sizes := []Size{
		{Label: "Huge", Width: 5000, Height: 3000, Source: "huge.jpg"},
		{Label: "Big", Width: 2500, Height: 1500, Source: "big.jpg"},
	}

	got, err := SelectSize(sizes, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// |2500 - 2048| = 452, |5000 - 2048| = 2952 -> "big.jpg" is closer
	if got.Source != "big.jpg" {
		t.Errorf("closest width: got %q, want %q", got.Source, "big.jpg")
	}
}

// ---------------------------------------------------------------------------
// SelectSizeByMaxDimension
// ---------------------------------------------------------------------------

func TestSelectSizeByMaxDimension(t *testing.T) {
	t.Run("empty returns error", func(t *testing.T) {
		_, err := SelectSizeByMaxDimension(nil, 1024)
		if err == nil {
			t.Fatal("expected error for nil sizes")
		}
		_, err = SelectSizeByMaxDimension([]Size{}, 1024)
		if err == nil {
			t.Fatal("expected error for empty sizes")
		}
	})

	t.Run("all within limit returns largest width", func(t *testing.T) {
		sizes := []Size{
			{Label: "Small", Width: 240, Height: 180, Source: "small.jpg"},
			{Label: "Medium", Width: 500, Height: 375, Source: "medium.jpg"},
			{Label: "Large", Width: 800, Height: 600, Source: "large.jpg"},
		}
		got, err := SelectSizeByMaxDimension(sizes, 1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Source != "large.jpg" {
			t.Errorf("got %q, want %q", got.Source, "large.jpg")
		}
	})

	t.Run("some within limit returns largest within", func(t *testing.T) {
		sizes := []Size{
			{Label: "Small", Width: 240, Height: 180, Source: "small.jpg"},
			{Label: "Medium", Width: 500, Height: 375, Source: "medium.jpg"},
			{Label: "Large", Width: 1024, Height: 768, Source: "large.jpg"},
			{Label: "Original", Width: 4000, Height: 3000, Source: "orig.jpg"},
		}
		got, err := SelectSizeByMaxDimension(sizes, 800)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Source != "medium.jpg" {
			t.Errorf("got %q, want %q", got.Source, "medium.jpg")
		}
	})

	t.Run("none within limit returns first element", func(t *testing.T) {
		sizes := []Size{
			{Label: "A", Width: 500, Height: 400, Source: "a.jpg"},
			{Label: "B", Width: 1000, Height: 800, Source: "b.jpg"},
		}
		got, err := SelectSizeByMaxDimension(sizes, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Code returns sizes[0] when nothing fits
		if got.Source != "a.jpg" {
			t.Errorf("got %q, want %q", got.Source, "a.jpg")
		}
	})

	t.Run("height exceeds but width fits", func(t *testing.T) {
		sizes := []Size{
			{Label: "Wide", Width: 400, Height: 2000, Source: "wide.jpg"},
			{Label: "Square", Width: 300, Height: 300, Source: "sq.jpg"},
		}
		got, err := SelectSizeByMaxDimension(sizes, 500)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Width 400 <= 500 but Height 2000 > 500 -> excluded
		if got.Source != "sq.jpg" {
			t.Errorf("got %q, want %q", got.Source, "sq.jpg")
		}
	})

	t.Run("exactly at limit", func(t *testing.T) {
		sizes := []Size{
			{Label: "Exact", Width: 800, Height: 600, Source: "exact.jpg"},
			{Label: "Under", Width: 400, Height: 300, Source: "under.jpg"},
		}
		got, err := SelectSizeByMaxDimension(sizes, 800)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Source != "exact.jpg" {
			t.Errorf("got %q, want %q", got.Source, "exact.jpg")
		}
	})
}

// ---------------------------------------------------------------------------
// base58Decode
// ---------------------------------------------------------------------------

func TestBase58Decode(t *testing.T) {
	t.Run("empty string returns zero", func(t *testing.T) {
		got, err := base58Decode("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("single character returns alphabet index", func(t *testing.T) {
		// '1' is index 0
		got, err := base58Decode("1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("'1' -> %d, want 0", got)
		}

		// '2' is index 1
		got, err = base58Decode("2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 1 {
			t.Errorf("'2' -> %d, want 1", got)
		}

		// 'Z' is the last character, index 57
		got, err = base58Decode("Z")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 57 {
			t.Errorf("'Z' -> %d, want 57", got)
		}
	})

	t.Run("two characters", func(t *testing.T) {
		// "11" = 0*58 + 0 = 0
		got, err := base58Decode("11")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("'11' -> %d, want 0", got)
		}

		// "21" = 1*58 + 0 = 58
		got, err = base58Decode("21")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 58 {
			t.Errorf("'21' -> %d, want 58", got)
		}
	})

	t.Run("known Flickr short URL value", func(t *testing.T) {
		got, err := base58Decode("2oFnQhB")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// This is the base58-decoded value for flic.kr/p/2oFnQhB
		if got != 52954144571 {
			t.Errorf("'2oFnQhB' -> %d, want 52954144571", got)
		}
	})

	t.Run("invalid characters rejected", func(t *testing.T) {
		for _, c := range []string{"0", "I", "O", "l"} {
			t.Run("char_"+c, func(t *testing.T) {
				_, err := base58Decode(c)
				if err == nil {
					t.Errorf("expected error for invalid base58 char %q", c)
				}
			})
		}
	})

	t.Run("invalid in middle of valid", func(t *testing.T) {
		_, err := base58Decode("2o0nQhB")
		if err == nil {
			t.Error("expected error for embedded '0'")
		}
	})
}

// ---------------------------------------------------------------------------
// DecodeShortURL
// ---------------------------------------------------------------------------

func TestDecodeShortURL(t *testing.T) {
	t.Run("flic.kr domain", func(t *testing.T) {
		got, err := DecodeShortURL("https://flic.kr/p/2oFnQhB")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "52954144571" {
			t.Errorf("got %q, want %q", got, "52954144571")
		}
	})

	t.Run("www.flickr.com domain", func(t *testing.T) {
		got, err := DecodeShortURL("https://www.flickr.com/p/2oFnQhB")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "52954144571" {
			t.Errorf("got %q, want %q", got, "52954144571")
		}
	})

	t.Run("flickr.com without www", func(t *testing.T) {
		got, err := DecodeShortURL("https://flickr.com/p/2oFnQhB")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "52954144571" {
			t.Errorf("got %q, want %q", got, "52954144571")
		}
	})

	t.Run("wrong host returns error", func(t *testing.T) {
		_, err := DecodeShortURL("https://example.com/p/2oFnQhB")
		if err == nil {
			t.Error("expected error for non-Flickr host")
		}
	})

	t.Run("missing /p/ prefix returns error", func(t *testing.T) {
		_, err := DecodeShortURL("https://flic.kr/photos/2oFnQhB")
		if err == nil {
			t.Error("expected error for missing /p/ prefix")
		}
	})

	t.Run("invalid base58 in path returns error", func(t *testing.T) {
		_, err := DecodeShortURL("https://flic.kr/p/0invalid")
		if err == nil {
			t.Error("expected error for invalid base58")
		}
	})

	t.Run("empty path segment returns error", func(t *testing.T) {
		_, err := DecodeShortURL("https://flic.kr/p/")
		if err == nil {
			t.Error("expected error for empty path after /p/")
		}
	})
}

// ---------------------------------------------------------------------------
// ResolvePhotoID
// ---------------------------------------------------------------------------

func TestResolvePhotoID(t *testing.T) {
	t.Run("bare numeric ID", func(t *testing.T) {
		got, err := ResolvePhotoID("12345")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "12345" {
			t.Errorf("got %q, want %q", got, "12345")
		}
	})

	t.Run("bare numeric with leading/trailing spaces", func(t *testing.T) {
		got, err := ResolvePhotoID("  98765  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "98765" {
			t.Errorf("got %q, want %q", got, "98765")
		}
	})

	t.Run("short URL via flic.kr", func(t *testing.T) {
		got, err := ResolvePhotoID("https://flic.kr/p/2oFnQhB")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "52954144571" {
			t.Errorf("got %q, want %q", got, "52954144571")
		}
	})

	t.Run("full flickr.com URL", func(t *testing.T) {
		got, err := ResolvePhotoID("https://www.flickr.com/photos/username/12345/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "12345" {
			t.Errorf("got %q, want %q", got, "12345")
		}
	})

	t.Run("full URL with extra path segments", func(t *testing.T) {
		got, err := ResolvePhotoID("https://www.flickr.com/photos/username/67890/in/album-123/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "67890" {
			t.Errorf("got %q, want %q", got, "67890")
		}
	})

	t.Run("full URL without trailing slash", func(t *testing.T) {
		got, err := ResolvePhotoID("https://www.flickr.com/photos/user/55555")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "55555" {
			t.Errorf("got %q, want %q", got, "55555")
		}
	})

	t.Run("unrecognized string returns error", func(t *testing.T) {
		_, err := ResolvePhotoID("not-a-photo-id")
		if err == nil {
			t.Error("expected error for unrecognized input")
		}
	})

	t.Run("non-numeric bare string returns error", func(t *testing.T) {
		_, err := ResolvePhotoID("abc123")
		if err == nil {
			t.Error("expected error for non-numeric bare string")
		}
	})

	t.Run("full URL with non-numeric photo segment returns error", func(t *testing.T) {
		_, err := ResolvePhotoID("https://www.flickr.com/photos/user/notanumber/")
		if err == nil {
			t.Error("expected error for non-numeric photo ID in URL")
		}
	})
}

// ---------------------------------------------------------------------------
// DeriveExtension
// ---------------------------------------------------------------------------

func TestDeriveExtension(t *testing.T) {
	tests := []struct {
		name           string
		sourceURL      string
		media          string
		originalFormat string
		want           string
	}{
		{"jpg from URL", "https://example.com/photo.jpg", "photo", "", "jpg"},
		{"jpeg normalised to jpg", "https://example.com/photo.jpeg", "photo", "", "jpg"},
		{"png from URL", "https://example.com/photo.png", "photo", "", "png"},
		{"gif from URL", "https://example.com/photo.gif", "photo", "", "gif"},
		{"webp from URL", "https://example.com/photo.webp", "photo", "", "webp"},
		{"jpg from URL with query params", "https://example.com/photo.jpg?size=large&farm=1", "photo", "", "jpg"},
		{"originalFormat jpeg normalised", "", "photo", "jpeg", "jpg"},
		{"originalFormat png", "", "photo", "png", "png"},
		{"originalFormat jpg", "", "photo", "jpg", "jpg"},
		{"video default mp4", "", "video", "", "mp4"},
		{"photo default jpg", "", "photo", "", "jpg"},
		{"empty everything defaults to jpg", "", "", "", "jpg"},
		{"URL takes priority over originalFormat", "https://example.com/image.png", "photo", "jpeg", "png"},
		{"mp4 from URL", "https://example.com/video.mp4", "video", "", "mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveExtension(tt.sourceURL, tt.media, tt.originalFormat)
			if got != tt.want {
				t.Errorf("DeriveExtension(%q, %q, %q) = %q, want %q",
					tt.sourceURL, tt.media, tt.originalFormat, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BestSizeURL
// ---------------------------------------------------------------------------

func TestBestSizeURL(t *testing.T) {
	t.Run("all fields set returns URLO", func(t *testing.T) {
		p := &PhotoListItem{
			URLO: "https://example.com/o.jpg",
			URLK: "https://example.com/k.jpg",
			URLL: "https://example.com/l.jpg",
			URLM: "https://example.com/m.jpg",
			URLS: "https://example.com/s.jpg",
		}
		got := BestSizeURL(p)
		if got != "https://example.com/o.jpg" {
			t.Errorf("got %q, want %q", got, "https://example.com/o.jpg")
		}
	})

	t.Run("only URLS set", func(t *testing.T) {
		p := &PhotoListItem{URLS: "https://example.com/s.jpg"}
		got := BestSizeURL(p)
		if got != "https://example.com/s.jpg" {
			t.Errorf("got %q, want %q", got, "https://example.com/s.jpg")
		}
	})

	t.Run("none set returns empty", func(t *testing.T) {
		p := &PhotoListItem{}
		got := BestSizeURL(p)
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("URLL + URLM returns URLL", func(t *testing.T) {
		p := &PhotoListItem{
			URLL: "https://example.com/l.jpg",
			URLM: "https://example.com/m.jpg",
		}
		got := BestSizeURL(p)
		if got != "https://example.com/l.jpg" {
			t.Errorf("got %q, want %q", got, "https://example.com/l.jpg")
		}
	})

	t.Run("URLK takes priority over URLL", func(t *testing.T) {
		p := &PhotoListItem{
			URLK: "https://example.com/k.jpg",
			URLL: "https://example.com/l.jpg",
			URLM: "https://example.com/m.jpg",
		}
		got := BestSizeURL(p)
		if got != "https://example.com/k.jpg" {
			t.Errorf("got %q, want %q", got, "https://example.com/k.jpg")
		}
	})

	t.Run("nil pointer panics are not expected but zero-value struct", func(t *testing.T) {
		// Verify zero-value PhotoListItem doesn't crash
		p := new(PhotoListItem)
		got := BestSizeURL(p)
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

// ---------------------------------------------------------------------------
// abs
// ---------------------------------------------------------------------------

func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want int
	}{
		{"positive", 42, 42},
		{"negative", -42, 42},
		{"zero", 0, 0},
		{"large positive", 1000000, 1000000},
		{"large negative", -1000000, 1000000},
		{"one", 1, 1},
		{"minus one", -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := abs(tt.x)
			if got != tt.want {
				t.Errorf("abs(%d) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}
