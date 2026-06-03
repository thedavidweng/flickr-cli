package flickr

import (
	"context"
	"encoding/json"
	"fmt"
)

// PageInfo holds pagination metadata from a Flickr response.
type PageInfo struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// PageResult holds a single page of results.
type PageResult[T any] struct {
	Info  PageInfo
	Items []T
}

// PageFetcher is a function that fetches a single page of results.
type PageFetcher[T any] func(ctx context.Context, page, perPage int) (PageResult[T], error)

// FetchAll fetches all pages of results using the provided fetcher.
func FetchAll[T any](ctx context.Context, perPage int, fetch PageFetcher[T], onPage func(PageInfo)) ([]T, error) {
	var allItems []T
	page := 1

	for {
		result, err := fetch(ctx, page, perPage)
		if err != nil {
			return allItems, fmt.Errorf("fetching page %d: %w", page, err)
		}

		allItems = append(allItems, result.Items...)

		if onPage != nil {
			onPage(result.Info)
		}

		if result.Info.Pages == 0 || page >= result.Info.Pages {
			break
		}
		page++
	}

	return allItems, nil
}

// extractPageInfo extracts PageInfo from a raw JSON response.
func extractPageInfo(raw json.RawMessage, infoKey string) (PageInfo, error) {
	var full map[string]json.RawMessage
	if err := json.Unmarshal(raw, &full); err != nil {
		return PageInfo{}, err
	}

	infoRaw, ok := full[infoKey]
	if !ok {
		return PageInfo{}, nil
	}

	var info PageInfo
	if err := json.Unmarshal(infoRaw, &info); err != nil {
		return PageInfo{}, err
	}
	return info, nil
}
