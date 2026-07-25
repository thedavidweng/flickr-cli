# ADR 0004: Lazy Walker for Memory-Efficient Pagination

## Status

Accepted

## Context

Flickr API endpoints return paginated results (`page`/`pages`/`perpage`/
`total`). A user's library can contain tens of thousands of photos across
hundreds of albums.

An eager helper that collects every page into a single slice before returning
allocates a large amount of memory at once, even if the caller only needs to
stream items one at a time (e.g. downloading, checksum verification, cache
sync).

## Decision

Provide a generic `Walker[T]` as the pagination mechanism. `Walker` fetches
pages on demand as items are consumed via a channel. Only one page is buffered
in memory at a time.

```go
w := flickr.NewWalker(ctx, 500, fetcher)
for item := range w.Items() {
    // process item
}
if err := w.Err(); err != nil { ... }
```

An eager `FetchAll` slice-collecting helper existed alongside `Walker` but had
no production callers and was removed. All pagination goes through `Walker`.

## Consequences

### Positive

- Constant memory for streaming operations regardless of collection size.
- A single pagination pattern in the package.
- Channel-based API integrates naturally with `for range` loops and
  context cancellation.

### Negative

- `Walker` errors surface after iteration completes (via `Err()`), not
  during the loop — callers must remember to check.
