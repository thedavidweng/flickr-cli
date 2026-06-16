package flickr

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
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

func BenchmarkFetchAll(b *testing.B) {
	b.ReportAllocs()

	itemsPerPage := 34 // 34 * 2 = 68; last page gets 100 - 68 = 32 items
	totalPages := 3

	fetcher := func(ctx context.Context, page, perPage int) (PageResult[int], error) {
		count := itemsPerPage
		if page == totalPages {
			count = 100 - itemsPerPage*(totalPages-1) // 32 items on last page
		}
		items := make([]int, count)
		return PageResult[int]{
			Info:  PageInfo{Page: page, Pages: totalPages, PerPage: perPage, Total: 100},
			Items: items,
		}, nil
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := FetchAll(context.Background(), itemsPerPage, fetcher, nil)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// --- Walker tests ---

func TestNewWalker(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		return PageResult[string]{}, nil
	}
	w := NewWalker(context.Background(), 10, fetcher)
	if w == nil {
		t.Fatal("NewWalker returned nil")
	}
	if w.perPage != 10 {
		t.Errorf("expected perPage 10, got %d", w.perPage)
	}
	if w.page != 1 {
		t.Errorf("expected page 1, got %d", w.page)
	}
	if w.done {
		t.Error("new walker should not be done")
	}
}

func TestWalkerSinglePage(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		return PageResult[string]{
			Info:  PageInfo{Page: 1, Pages: 1, PerPage: perPage},
			Items: []string{"a", "b", "c"},
		}, nil
	}

	w := NewWalker(context.Background(), 100, fetcher)
	var got []string
	for item := range w.Items() {
		got = append(got, item)
	}
	if err := w.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("unexpected items: %v", got)
	}
}

func TestWalkerMultiplePages(t *testing.T) {
	var mu sync.Mutex
	fetchedPages := make(map[int]bool)
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[int], error) {
		mu.Lock()
		fetchedPages[page] = true
		mu.Unlock()
		switch page {
		case 1:
			return PageResult[int]{
				Info:  PageInfo{Page: 1, Pages: 3, PerPage: perPage},
				Items: []int{1, 2},
			}, nil
		case 2:
			return PageResult[int]{
				Info:  PageInfo{Page: 2, Pages: 3, PerPage: perPage},
				Items: []int{3, 4},
			}, nil
		case 3:
			return PageResult[int]{
				Info:  PageInfo{Page: 3, Pages: 3, PerPage: perPage},
				Items: []int{5, 6},
			}, nil
		default:
			return PageResult[int]{
				Info: PageInfo{Page: page, Pages: 3, PerPage: perPage},
			}, nil
		}
	}

	w := NewWalker(context.Background(), 2, fetcher)
	var got []int
	for item := range w.Items() {
		got = append(got, item)
	}
	if err := w.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("expected 6 items, got %d: %v", len(got), got)
	}
	for i, v := range got {
		if v != i+1 {
			t.Errorf("item %d: expected %d, got %d", i, i+1, v)
		}
	}

	// Verify all 3 pages were fetched.
	mu.Lock()
	defer mu.Unlock()
	for p := 1; p <= 3; p++ {
		if !fetchedPages[p] {
			t.Errorf("page %d should have been fetched", p)
		}
	}
}

func TestWalkerEmptyResult(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		return PageResult[string]{
			Info: PageInfo{Page: 1, Pages: 0, PerPage: perPage},
		}, nil
	}

	w := NewWalker(context.Background(), 100, fetcher)
	var count int
	for range w.Items() {
		count++
	}
	if err := w.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 items, got %d", count)
	}
}

func TestWalkerFetcherError(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[int], error) {
		if page == 1 {
			return PageResult[int]{
				Info:  PageInfo{Page: 1, Pages: 2, PerPage: perPage},
				Items: []int{1, 2},
			}, nil
		}
		return PageResult[int]{}, errors.New("network timeout")
	}

	w := NewWalker(context.Background(), 100, fetcher)
	var got []int
	for item := range w.Items() {
		got = append(got, item)
	}
	err := w.Err()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching page 2") {
		t.Errorf("error %q should contain 'fetching page 2'", err.Error())
	}
	if len(got) != 2 {
		t.Errorf("expected 2 partial items, got %d", len(got))
	}
}

func TestWalkerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	fetcher := func(ctx context.Context, page, perPage int) (PageResult[int], error) {
		return PageResult[int]{
			Info:  PageInfo{Page: page, Pages: 3, PerPage: perPage},
			Items: []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		}, nil
	}

	w := NewWalker(ctx, 5, fetcher)
	ch := w.Items()

	// Consume one item, then cancel.
	<-ch
	cancel()

	// Wait for the walker goroutine to set w.err WITHOUT consuming from ch.
	// Since ch is unbuffered, not reading forces the goroutine's
	// select { case ch <- item: case <-ctx.Done(): } to pick ctx.Done()
	// deterministically (ch<- blocks when nobody reads).
	deadline := time.After(5 * time.Second)
	for {
		if w.Err() != nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for walker to detect cancellation")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	if w.Err() != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", w.Err())
	}

	// Drain ch to let the goroutine finish and avoid goroutine leak.
	for range ch {
	}
}

func TestWalkerManyItemsAcrossPages(t *testing.T) {
	fetcher := func(ctx context.Context, page, perPage int) (PageResult[string], error) {
		switch page {
		case 1:
			return PageResult[string]{
				Info:  PageInfo{Page: 1, Pages: 3, PerPage: perPage},
				Items: []string{"a1", "a2"},
			}, nil
		case 2:
			return PageResult[string]{
				Info:  PageInfo{Page: 2, Pages: 3, PerPage: perPage},
				Items: []string{"b1", "b2"},
			}, nil
		case 3:
			return PageResult[string]{
				Info:  PageInfo{Page: 3, Pages: 3, PerPage: perPage},
				Items: []string{"c1", "c2"},
			}, nil
		default:
			return PageResult[string]{
				Info: PageInfo{Page: page, Pages: 3, PerPage: perPage},
			}, nil
		}
	}

	w := NewWalker(context.Background(), 2, fetcher)
	var got []string
	for item := range w.Items() {
		got = append(got, item)
	}
	if err := w.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"a1", "a2", "b1", "b2", "c1", "c2"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(got))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("item %d: expected %q, got %q", i, expected[i], v)
		}
	}
}
