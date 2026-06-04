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
// Currently only used in tests; kept as a utility for future use.
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

// Walker lazily iterates through paginated Flickr API results.
// It fetches pages on demand as items are consumed, similar to
// the Python library's Walker pattern. This is memory-efficient
// for large collections because only one page is buffered at a time.
//
// Usage:
//
//	w := flickr.NewWalker(ctx, 500, fetcher)
//	for item := range w.Items() {
//	    // process item
//	}
//	if err := w.Err(); err != nil { ... }
type Walker[T any] struct {
	ctx      context.Context
	fetch    PageFetcher[T]
	perPage  int
	page     int
	pages    int
	buf      []T
	bufIdx   int
	err      error
	done     bool
}

// NewWalker creates a Walker that lazily paginates through results.
func NewWalker[T any](ctx context.Context, perPage int, fetch PageFetcher[T]) *Walker[T] {
	return &Walker[T]{
		ctx:     ctx,
		fetch:   fetch,
		perPage: perPage,
		page:    1,
	}
}

// next fetches the next page if the buffer is exhausted.
func (w *Walker[T]) next() bool {
	if w.done {
		return false
	}
	if w.bufIdx < len(w.buf) {
		return true
	}

	result, err := w.fetch(w.ctx, w.page, w.perPage)
	if err != nil {
		w.err = fmt.Errorf("fetching page %d: %w", w.page, err)
		w.done = true
		return false
	}

	w.buf = result.Items
	w.bufIdx = 0
	w.pages = result.Info.Pages

	if len(w.buf) == 0 || w.page >= w.pages {
		w.done = true
		return len(w.buf) > 0
	}
	w.page++
	return true
}

// Items returns a channel that yields items lazily.
// The channel is closed when all pages are exhausted or an error occurs.
func (w *Walker[T]) Items() <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for w.next() {
			for w.bufIdx < len(w.buf) {
				item := w.buf[w.bufIdx]
				w.bufIdx++
				select {
				case ch <- item:
				case <-w.ctx.Done():
					w.err = w.ctx.Err()
					return
				}
			}
		}
	}()
	return ch
}

// Err returns the first error encountered during iteration, if any.
func (w *Walker[T]) Err() error {
	return w.err
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
