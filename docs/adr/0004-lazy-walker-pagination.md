# ADR 0004: Lazy Walker for Memory-Efficient Pagination

## Status

Accepted

## Context

Flickr API endpoints return paginated results (`page`/`pages`/`perpage`/
`total`). A user's library can contain tens of thousands of photos across
hundreds of albums.

`FetchAll` collects every page into a single slice before returning. For
large collections this allocates a large amount of memory at once, even if
the caller only needs to stream items one at a time (e.g. downloading,
checksum verification, cache sync).

## Decision

Provide a generic `Walker[T]` alongside `FetchAll`. `Walker` fetches pages on
demand as items are consumed via a channel. Only one page is buffered in
memory at a time.

```go
w := flickr.NewWalker(ctx, 500, fetcher)
for item := range w.Items() {
    // process item
}
if err := w.Err(); err != nil { ... }
```

`FetchAll` remains for callers that need the full slice; `Walker` is for
streaming and large collections.

## Consequences

### Positive

- Constant memory for streaming operations regardless of collection size.
- Caller chooses the right tool: `FetchAll` for small sets, `Walker` for
  large or streaming.
- Channel-based API integrates naturally with `for range` loops and
  context cancellation.

### Negative

- Two pagination patterns exist in the same package.
- `Walker` errors surface after iteration completes (via `Err()`), not
  during the loop — callers must remember to check.
