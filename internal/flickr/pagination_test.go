package flickr

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFetchAll(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		if page > 2 {
			return PageResult[string]{
				Info: PageInfo{Page: page, Pages: 2, PerPage: perPage},
			}, nil
		}
		return PageResult[string]{
			Info:  PageInfo{Page: page, Pages: 2, PerPage: perPage},
			Items: []string{"item1", "item2"},
		}, nil
	}

	items, err := FetchAll(context.Background(), 100, fetcher, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 4 {
		t.Errorf("expected 4 items, got %d", len(items))
	}
}

func TestFetchAllEmpty(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		return PageResult[string]{
			Info: PageInfo{Page: 1, Pages: 0, PerPage: perPage},
		}, nil
	}

	items, err := FetchAll(context.Background(), 100, fetcher, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestFetchAllWithCallback(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		return PageResult[string]{
			Info:  PageInfo{Page: page, Pages: 1, PerPage: perPage},
			Items: []string{"item1"},
		}, nil
	}

	var callbackPage PageInfo
	callback := func(info PageInfo) {
		callbackPage = info
	}

	_, err := FetchAll(context.Background(), 100, fetcher, callback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callbackPage.Page != 1 {
		t.Errorf("expected page 1, got %d", callbackPage.Page)
	}
}

func TestExtractPageInfo(t *testing.T) {
	// The extractPageInfo function expects a wrapper object with the info key
	raw := json.RawMessage(`{"photos":{"page":1,"pages":5,"per_page":100,"total":500}}`)
	info, err := extractPageInfo(raw, "photos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Page != 1 {
		t.Errorf("expected page 1, got %d", info.Page)
	}
	if info.Pages != 5 {
		t.Errorf("expected pages 5, got %d", info.Pages)
	}
}

func TestExtractPageInfoMissingKey(t *testing.T) {
	raw := json.RawMessage(`{"other":{"page":1}}`)
	info, err := extractPageInfo(raw, "photos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Page != 0 {
		t.Errorf("expected page 0 for missing key, got %d", info.Page)
	}
}
