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
